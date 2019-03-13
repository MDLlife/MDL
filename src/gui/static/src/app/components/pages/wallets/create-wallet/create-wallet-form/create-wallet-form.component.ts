import { Component, OnInit, OnDestroy, Input } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { ISubscription } from 'rxjs/Subscription';
import { ApiService } from '../../../../../services/api.service';

export class FormData {
  label: string;
  seed: string;
  password: string;
}

@Component({
  selector: 'app-create-wallet-form',
  templateUrl: './create-wallet-form.component.html',
  styleUrls: ['./create-wallet-form.component.scss'],
})
export class CreateWalletFormComponent implements OnInit, OnDestroy {
  @Input() create: boolean;
  @Input() whiteText: boolean;
  @Input() onboarding: boolean;

  form: FormGroup;
  normalSeed = false;
  customSeedAccepted = false;
  encrypt = true;

  private statusSubscription: ISubscription;

  constructor(
    private apiService: ApiService,
  ) { }

  ngOnInit() {
    if (!this.onboarding) {
      this.initForm();
    } else {
      this.initForm(false, null);
    }
  }

  ngOnDestroy() {
    this.statusSubscription.unsubscribe();
  }

  get isValid(): boolean {
    return this.form.valid && (this.normalSeed || this.customSeedAccepted);
  }

  onCustomSeedAcceptance(event) {
    this.customSeedAccepted = event.checked;
  }

  setEncrypt(event) {
    this.encrypt = event.checked;
    this.form.updateValueAndValidity();
  }

  getData(): FormData {
    return {
      label: this.form.value.label,
      seed: this.form.value.seed,
      password: !this.onboarding && this.encrypt ? this.form.value.password : null,
    };
  }

  initForm(create: boolean = null, data: Object = null) {
    create = create !== null ? create : this.create;

    const validators = [];
    if (create) {
      validators.push(this.seedMatchValidator.bind(this));
    }
    if (!this.onboarding) {
      validators.push(this.validatePasswords.bind(this));
    }

    this.form = new FormGroup({}, validators);
    this.form.addControl('label', new FormControl(data ? data['label'] : '', [Validators.required]));
    this.form.addControl('seed', new FormControl(data ? data['seed'] : '', [Validators.required]));
    this.form.addControl('confirm_seed', new FormControl(data ? data['seed'] : ''));
    this.form.addControl('password', new FormControl());
    this.form.addControl('confirm_password', new FormControl());

    if (create && !data) {
      this.generateSeed(128);
    }

    if (data) {
      this.normalSeed = this.isNormalSeed(data['seed']);
      this.customSeedAccepted = true;
    }

    if (this.statusSubscription && !this.statusSubscription.closed) {
      this.statusSubscription.unsubscribe();
    }
    this.statusSubscription = this.form.statusChanges.subscribe(() => {
      this.customSeedAccepted = false;
      this.normalSeed = this.isNormalSeed(this.form.get('seed').value);
    });
  }

  generateSeed(entropy: number) {
    this.apiService.generateSeed(entropy).subscribe(seed => this.form.get('seed').setValue(seed));
  }

  private isNormalSeed(seed: string): boolean {
    const processedSeed = seed.replace(/\r?\n|\r/g, ' ').replace(/ +/g, ' ').trim();
    if (seed !== processedSeed) {
      return false;
    }

    const NumberOfWords = seed.split(' ').length;
    if (NumberOfWords !== 12 && NumberOfWords !== 24) {
      return false;
    }

    if (!(/^[a-z\s]*$/).test(seed)) {
      return false;
    }

    return true;
  }

  private validatePasswords() {
    if (this.encrypt && this.form && this.form.get('password') && this.form.get('confirm_password')) {
      if (this.form.get('password').value) {
        if (this.form.get('password').value !== this.form.get('confirm_password').value) {
          return { NotEqual: true };
        }
      } else {
        return { Required: true };
      }
    }

    return null;
  }

  private seedMatchValidator() {
    if (this.form && this.form.get('seed') && this.form.get('confirm_seed')) {
      return this.form.get('seed').value === this.form.get('confirm_seed').value ? null : { NotEqual: true };
    } else {
      return { NotEqual: true };
    }
  }
}
