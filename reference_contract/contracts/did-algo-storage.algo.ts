// eslint-disable-next-line import/no-extraneous-dependencies
import { assert, BoxMap, bytes, clone, Contract, GlobalState, gtxn, itxn, Txn, Uint64, uint64 } from '@algorandfoundation/algorand-typescript';
import { Address, Byte} from '@algorandfoundation/algorand-typescript/arc4';

/** Metadata about DID Document data */
type Metadata = {
  /** start - The index of the box at which the data starts */
  start: uint64,

  /** end - The index of the box at which the data ends */
  end: uint64,
  /** status - 0 if uploading, 1 if ready, 2 if deleting */
  status: Byte

  /** The size of the last box */
  endSize: uint64,

  /**
   * The index of the last box that was deleted. Used to ensure all boxes are deleted in order
   * To prevent any missed boxes (thus missed MBR refund)
   */
  lastDeleted: uint64
};

/** Indicates the data is still being uploaded */
const UPLOADING: Byte = new Byte(0);

/** Indicates the data is done uploading and can be safely read */
const READY: Byte = new Byte(1);

/** Indicates the data is currently being deleted */
const DELETING: Byte = new Byte(2);

const COST_PER_BYTE = 400;
const COST_PER_BOX = 2500;
const MAX_BOX_SIZE = 32768;

export class DIDAlgoStorage extends Contract {
  dataBoxes = BoxMap<uint64, bytes>({ keyPrefix: '' });
  metadata = BoxMap<Address, Metadata>({ keyPrefix: '' });
  currentIndex = GlobalState<uint64>({ initialValue: Uint64(0) });

  /**
 *
 * Allocate boxes to begin data upload process
 *
 * @param pubKey The pubkey of the DID
 * @param numBoxes The number of boxes that the data will take up
 * @param endBoxSize The size of the last box
 * @param mbrPayment Payment from the uploader to cover the box MBR
 */
  startUpload(
    pubKey: Address,
    numBoxes: uint64,
    endBoxSize: uint64,
    mbrPayment: gtxn.PaymentTxn,
  ): void {
    assert(Txn.sender === Txn.applicationId.creator);

    const startBox: uint64 = this.currentIndex.value;

    const endBox: uint64 = startBox + numBoxes - 1;

    const metadata: Metadata = {
      start: startBox, end: endBox, status: UPLOADING, endSize: endBoxSize, lastDeleted: 0,
    };

    assert(!this.metadata(pubKey).exists);

    this.metadata(pubKey).value = clone(metadata);

    this.currentIndex.value = endBox + 1;

    const totalCost: uint64 = numBoxes * COST_PER_BOX // cost of data boxes
      + (numBoxes - 1) * MAX_BOX_SIZE * COST_PER_BYTE // cost of data
      + numBoxes * 8 * COST_PER_BYTE // cost of data keys
      + endBoxSize * COST_PER_BYTE // cost of last data box
      + COST_PER_BOX + (8 + 8 + 1 + 8 + 32 + 8) * COST_PER_BYTE; // cost of metadata box

    assert(mbrPayment.amount === totalCost);
  }


  /**
   *
   * Upload data to a specific offset in a box
   *
   * @param pubKey The pubkey of the DID
   * @param boxIndex The index of the box to upload the given chunk of data to
   * @param offset The offset within the box to start writing the data
   * @param data The data to write
   */
  upload(pubKey: Address, boxIndex: uint64, offset: uint64, data: bytes): void {
    assert(Txn.sender === Txn.applicationId.creator);

    const metadata = clone(this.metadata(pubKey).value);
    assert(metadata.status === UPLOADING);
    
    assert(metadata.start <= boxIndex && boxIndex <= metadata.end);

    if (offset === 0) {
      this.dataBoxes(boxIndex).create({
        size: boxIndex === metadata.end ? metadata.endSize : MAX_BOX_SIZE
      })
    }

    this.dataBoxes(boxIndex).replace(offset, data);
  }

  /**
   *
   * Mark uploading as false
   *
   * @param pubKey The address of the DID
   */
  finishUpload(pubKey: Address): void {
    assert(Txn.sender === Txn.applicationId.creator);
    this.metadata(pubKey).value.status = READY;
  }

  /**
   * Starts the deletion process for the data associated with a DID
   *
   * @param pubKey The address of the DID
   */
  startDelete(pubKey: Address): void {
    assert(Txn.sender === Txn.applicationId.creator);

    const metadata = clone(this.metadata(pubKey).value);
    assert(metadata.status === READY);

    metadata.status = DELETING;
    
    // Since variables are assigned by value, we need to clone before modifying so we have the right value
    this.metadata(pubKey).value = clone(metadata);
  }

  /**
   * Deletes a box of data
   *
   * @param pubKey The address of the DID
   * @param boxIndex The index of the box to delete
   */
  deleteData(pubKey: Address, boxIndex: uint64): void {
    assert(Txn.sender === Txn.applicationId.creator);

    const metadata = clone(this.metadata(pubKey).value);
    assert(metadata.status === DELETING);
    assert(metadata.start <= boxIndex && boxIndex <= metadata.end);

    if (boxIndex !== metadata.start) assert(metadata.lastDeleted === boxIndex - 1);

    const preMBR = Txn.applicationId.address.minBalance;

    this.dataBoxes(boxIndex).delete();

    // If this is the last box, delete the metadata
    if (boxIndex === metadata.end) {
      this.metadata(pubKey).delete();
    // Otherwise, update the last deleted index
    } else {
      metadata.lastDeleted = boxIndex;
      this.metadata(pubKey).value = clone(metadata);
    }

    itxn.payment({
      amount: preMBR - Txn.applicationId.address.minBalance,
      receiver: Txn.sender,
    }).submit();
  }

  /**
   * Allow the contract to be updated by the creator
   */
  updateApplication(): void {
    assert(Txn.sender === Txn.applicationId.creator);
  }

  /**
   * Dummy function to add extra box references for deleteData.
   * Boxes are 32k, but a single app call can only inlcude enough references to read/write 8k
   * at a time. Thus when a box is deleted, we need to add additional dummy calls with box
   * references to increase the total read/write budget to 32k.
   */
  dummy(): void {}
}
