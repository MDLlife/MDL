import { Component, OnDestroy } from '@angular/core';
import { NavBarService } from '../../../services/nav-bar.service';
import { ISubscription } from 'rxjs/Subscription';
import { DoubleButtonActive } from '../../layout/double-button/double-button.component';

@Component({
  selector: 'app-send-mdl',
  templateUrl: './send-mdl.component.html',
  styleUrls: ['./send-mdl.component.scss'],
})
export class SendMDLComponent implements OnDestroy {
  showForm = true;
  formData: any;
  activeForm: DoubleButtonActive;
  activeForms = DoubleButtonActive;

  private subscription: ISubscription;

  constructor(
    private navbarService: NavBarService,
  ) {
    this.subscription = navbarService.activeComponent.subscribe(value => {
      this.activeForm = value;
      this.formData = null;
    });
  }

  ngOnDestroy() {
    this.subscription.unsubscribe();
  }

  onFormSubmitted(data) {
    this.formData = data;
    this.showForm = false;
  }

  onBack(deleteFormData) {
    if (deleteFormData) {
      this.formData = null;
    }

    this.showForm = true;
  }

  get transaction() {
    const transaction = this.formData.transaction;

    transaction.wallet = this.formData.form.wallet;
    transaction.from = this.formData.form.wallet.label;
    transaction.to = this.formData.to;
    transaction.balance = this.formData.amount;

    return transaction;
  }
}
