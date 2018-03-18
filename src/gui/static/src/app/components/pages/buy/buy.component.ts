import { Component, ViewChild } from '@angular/core';
import { PurchaseService } from '../../../services/purchase.service';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { WalletService } from '../../../services/wallet.service';
import { Address, PurchaseOrder, Wallet } from '../../../app.datatypes';
import { MatSnackBar } from '@angular/material/snack-bar';
import { ButtonComponent } from '../../layout/button/button.component';
import { Observable } from 'rxjs/Observable';

@Component({
  selector: 'app-buy',
  templateUrl: './buy.component.html',
  styleUrls: ['./buy.component.scss']
})
export class BuyComponent {
  @ViewChild('button') button: ButtonComponent;

  address: Address;
  config: any;
  supported: string[];
  methods: any;
  form: FormGroup;
  order: PurchaseOrder;
  wallets: Wallet[];

  constructor(
    private formBuilder: FormBuilder,
    private purchaseService: PurchaseService,
    private snackBar: MatSnackBar,
    private walletService: WalletService,
  ) {}

  ngOnInit() {
    this.initForm();
    this.loadData();
  }

  checkStatus() {
    this.button.setLoading();
    this.purchaseService.scan(this.order.recipient_address).first().subscribe(
      response => {
        this.button.setSuccess();
        this.order.status = response.status;
      },
      error => this.button.setError(error)
    );
  }

  removeOrder() {
    window.localStorage.removeItem('purchaseOrder');
    this.order = null;
    this.form.controls.coin_type.setValue('', { emitEvent: false });
  }

  private initForm() {
    this.form = this.formBuilder.group({
      wallet: ['', Validators.required],
      coin_type: ['', Validators.required],
    });

    this.form.controls.wallet.valueChanges.subscribe(filename => {
      if (this.form.value.coin_type === '') return;
      const wallet = this.wallets.find(wallet => wallet.filename === filename);
      console.log('changing wallet value', filename);
      this.purchaseService.generate(wallet, this.form.value.coin_type).subscribe(
        order => this.saveData(order),
        err => {
          this.snackBar.open(err._body);
        }
      );
    })
    this.form.controls.coin_type.valueChanges.subscribe(type => {
      if (type === '') return;
      if (this.form.value.wallet === '') return;
      const wallet = this.wallets.find(wallet => wallet.filename === this.form.value.wallet);
      this.purchaseService.generate(wallet, type).subscribe(
        order => this.saveData(order),
        err => {
          this.snackBar.open(err._body);
        }
      );
    })
  }

  private loadConfig() {
    this.purchaseService.config()
      .filter(config => !!config && !!config.mdl_btc_exchange_rate)
      .first()
      .subscribe(config => {
        this.config = config;
        this.supported = config.supported;
      });
  }

  private loadData() {
    this.loadConfig();
    this.loadOrder();

    this.walletService.all().subscribe(wallets => {
      this.wallets = wallets;

      if (this.order) {
        this.form.controls.wallet.setValue(this.order.filename, { emitEvent: false });
        this.form.controls.coin_type.setValue(this.order.coin_type, { emitEvent: false });
      }
    });
  }

  private loadOrder() {
    if (this.form.value.coin_type === '' || this.form.value.coin_type === '') return;

    const order: PurchaseOrder = JSON.parse(window.localStorage.getItem('purchaseOrder'));
    if (order) {
      this.order = order;
      this.updateOrder();
    }
  }

  private saveData(order: PurchaseOrder) {
    this.order = order;
    window.localStorage.setItem('purchaseOrder', JSON.stringify(order));
  }

  private updateOrder() {
    this.purchaseService.scan(this.order.recipient_address).first().subscribe(
      response => this.order.status = response.status,
      error => console.log(error)
    );
  }

  public currentCoinPrice() {
    switch (this.form.value.coin_type) {
      case "BTC": return this.config.mdl_btc_exchange_rate;
      case "ETH": return this.config.mdl_eth_exchange_rate;
      case "SKY": return this.config.mdl_sky_exchange_rate;
      case "WAVES": return this.config.mdl_waves_exchange_rate;
    }
    return "???";
  }
}
