import { Component, Inject, OnDestroy, OnInit } from '@angular/core';
import { MD_DIALOG_DATA, MdDialogRef } from '@angular/material';
import { Transaction } from '../../../../app.datatypes';

@Component({
  selector: 'app-transaction-detail',
  templateUrl: './transaction-detail.component.html',
  styleUrls: ['./transaction-detail.component.scss']
})
export class TransactionDetailComponent implements OnInit, OnDestroy {

  price: number;

  constructor(
    @Inject(MD_DIALOG_DATA) public transaction: Transaction,
    public dialogRef: MdDialogRef<TransactionDetailComponent>,
  ) {}

  ngOnInit() {

  }

  ngOnDestroy() {

  }

  closePopup() {
    this.dialogRef.close();
  }

  showOutput(output) {
    return !this.transaction.inputs.find(input => input.owner === output.dst);
  }
}
