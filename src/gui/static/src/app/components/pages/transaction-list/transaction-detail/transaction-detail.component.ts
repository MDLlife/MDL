import { Component, Inject, OnDestroy, OnInit } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { Transaction } from '../../../../app.datatypes';

@Component({
  selector: 'app-transaction-detail',
  templateUrl: './transaction-detail.component.html',
  styleUrls: ['./transaction-detail.component.scss']
})
export class TransactionDetailComponent implements OnInit, OnDestroy {

  price: number;

  constructor(
    @Inject(MAT_DIALOG_DATA) public transaction: Transaction,
    public dialogRef: MatDialogRef<TransactionDetailComponent>,
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
