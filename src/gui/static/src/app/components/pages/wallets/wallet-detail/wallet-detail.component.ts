import { Component, Input, OnDestroy, OnInit } from '@angular/core';
import { Wallet, ConfirmationData } from '../../../../app.datatypes';
import { WalletService } from '../../../../services/wallet.service';
import { MatDialog, MatDialogConfig, MatDialogRef } from '@angular/material/dialog';
import { ChangeNameComponent } from '../change-name/change-name.component';
import { QrCodeComponent } from '../../../layout/qr-code/qr-code.component';
import { PasswordDialogComponent } from '../../../layout/password-dialog/password-dialog.component';
import { MatSnackBar } from '@angular/material';
import { showSnackbarError, getHardwareWalletErrorMsg } from '../../../../utils/errors';
import { NumberOfAddressesComponent } from '../number-of-addresses/number-of-addresses';
import { TranslateService } from '@ngx-translate/core';
import { HwWalletService } from '../../../../services/hw-wallet.service';
import { Observable } from 'rxjs/Observable';
import { showConfirmationModal } from '../../../../utils';
import { AppConfig } from '../../../../app.config';
import { Router } from '@angular/router';
import { HwConfirmAddressDialogComponent, AddressConfirmationParams } from '../../../layout/hardware-wallet/hw-confirm-address-dialog/hw-confirm-address-dialog.component';

@Component({
  selector: 'app-wallet-detail',
  templateUrl: './wallet-detail.component.html',
  styleUrls: ['./wallet-detail.component.scss'],
})
export class WalletDetailComponent implements OnDestroy {
  @Input() wallet: Wallet;

  creatingAddress = false;

  private howManyAddresses: number;

  constructor(
    private dialog: MatDialog,
    private walletService: WalletService,
    private snackbar: MatSnackBar,
    private hwWalletService: HwWalletService,
    private translateService: TranslateService,
    private router: Router,
  ) { }

  ngOnDestroy() {
    this.snackbar.dismiss();
  }

  editWallet() {
    const config = new MatDialogConfig();
    config.width = '566px';
    config.data = this.wallet;
    this.dialog.open(ChangeNameComponent, config);
  }

  newAddress() {
    if (this.creatingAddress) {
      return;
    }

    if (this.wallet.isHardware && this.wallet.addresses.length >= AppConfig.maxHardwareWalletAddresses) {
      const confirmationData: ConfirmationData = {
        text: 'wallet.max-hardware-wallets-error',
        headerText: 'errors.error',
        confirmButtonText: 'confirmation.close',
      };
      showConfirmationModal(this.dialog, confirmationData);

      return;
    }

    this.snackbar.dismiss();

    if (!this.wallet.isHardware) {
      const config = new MatDialogConfig();
      config.width = '566px';
      this.dialog.open(NumberOfAddressesComponent, config).afterClosed()
        .subscribe(response => {
          if (response) {
            this.howManyAddresses = response;

            let lastWithBalance = 0;
            this.wallet.addresses.forEach((address, i) => {
              if (address.coins.isGreaterThan(0)) {
                lastWithBalance = i;
              }
            });

            if ((this.wallet.addresses.length - (lastWithBalance + 1)) + response < 20) {
              this.continueNewAddress();
            } else {
              const confirmationData: ConfirmationData = {
                text: 'wallet.add-many-confirmation',
                headerText: 'confirmation.header-text',
                confirmButtonText: 'confirmation.confirm-button',
                cancelButtonText: 'confirmation.cancel-button',
              };

              showConfirmationModal(this.dialog, confirmationData).afterClosed().subscribe(confirmationResult => {
                if (confirmationResult) {
                  this.continueNewAddress();
                }
              });
            }
          }
        });
    } else {
      this.howManyAddresses = 1;
      this.continueNewAddress();
    }
  }

  toggleEmpty() {
    this.wallet.hideEmpty = !this.wallet.hideEmpty;
  }

  deleteWallet() {
    const confirmationData: ConfirmationData = {
      text: this.translateService.instant('wallet.delete-confirmation', {name: this.wallet.label}),
      headerText: 'confirmation.header-text',
      checkboxText: 'wallet.delete-confirmation-check',
      confirmButtonText: 'confirmation.confirm-button',
      cancelButtonText: 'confirmation.cancel-button',
    };

    showConfirmationModal(this.dialog, confirmationData).afterClosed().subscribe(confirmationResult => {
      if (confirmationResult) {
        this.walletService.deleteHardwareWallet(this.wallet).subscribe(result => {
          if (result) {
            this.walletService.all().first().subscribe(wallets => {
              if (wallets.length === 0) {
                setTimeout(() => this.router.navigate(['/wizard']), 500);
              }
            });
          }
        });
      }
    });
  }

  toggleEncryption() {
    const config = new MatDialogConfig();
    config.data = {
      confirm: !this.wallet.encrypted,
      title: this.wallet.encrypted ? 'wallet.decrypt' : 'wallet.encrypt',
    };

    if (!this.wallet.encrypted) {
      config.data['description'] = 'wallet.new.encrypt-warning';
    } else {
      config.data['description'] = 'wallet.decrypt-warning';
      config.data['warning'] = true;
      config.data['wallet'] = this.wallet;
    }

    this.dialog.open(PasswordDialogComponent, config).componentInstance.passwordSubmit
      .subscribe(passwordDialog => {
        this.walletService.toggleEncryption(this.wallet, passwordDialog.password).subscribe(() => {
          passwordDialog.close();
        }, e => passwordDialog.error(e));
      });
  }

  confirmAddress(address, addressIndex, showCompleteConfirmation) {
    this.hwWalletService.checkIfCorrectHwConnected(this.wallet.addresses[0].address).subscribe(response => {
      const data = new AddressConfirmationParams();
      data.address = address;
      data.addressIndex = addressIndex;
      data.showCompleteConfirmation = showCompleteConfirmation;

      const config = new MatDialogConfig();
      config.width = '566px';
      config.autoFocus = false;
      config.data = data;
      this.dialog.open(HwConfirmAddressDialogComponent, config);
    }, err => {
      showSnackbarError(this.snackbar, getHardwareWalletErrorMsg(this.hwWalletService, this.translateService, err));
    });
  }

  copyAddress(event, address, duration = 500) {
    event.stopPropagation();

    if (address.copying) {
      return;
    }

    const selBox = document.createElement('textarea');

    selBox.style.position = 'fixed';
    selBox.style.left = '0';
    selBox.style.top = '0';
    selBox.style.opacity = '0';
    selBox.value = address.address;

    document.body.appendChild(selBox);
    selBox.focus();
    selBox.select();

    document.execCommand('copy');
    document.body.removeChild(selBox);

    address.copying = true;

    setTimeout(function() {
      address.copying = false;
    }, duration);
  }

  showQrCode(event, address: string) {
    event.stopPropagation();

    const config = new MatDialogConfig();
    config.data = { address };
    this.dialog.open(QrCodeComponent, config);
  }

  private continueNewAddress() {
    this.creatingAddress = true;

    if (!this.wallet.isHardware && this.wallet.encrypted) {
      const config = new MatDialogConfig();
      config.data = {
        wallet: this.wallet,
      };

      const dialogRef = this.dialog.open(PasswordDialogComponent, config);
      dialogRef.afterClosed().subscribe(() => this.creatingAddress = false);
      dialogRef.componentInstance.passwordSubmit
        .subscribe(passwordDialog => {
          this.walletService.addAddress(this.wallet, this.howManyAddresses, passwordDialog.password)
            .subscribe(() => passwordDialog.close(), error => passwordDialog.error(error));
        });
    } else {

      let procedure: Observable<any>;

      if (this.wallet.isHardware ) {
        procedure = this.hwWalletService.checkIfCorrectHwConnected(this.wallet.addresses[0].address).flatMap(
          () => this.walletService.addAddress(this.wallet, this.howManyAddresses),
        );
      } else {
        procedure = this.walletService.addAddress(this.wallet, this.howManyAddresses);
      }

      procedure.subscribe(() => this.creatingAddress = false,
        err => {
          if (!this.wallet.isHardware ) {
            showSnackbarError(this.snackbar, err);
          } else {
            showSnackbarError(this.snackbar, getHardwareWalletErrorMsg(this.hwWalletService, this.translateService, err));
          }
          this.creatingAddress = false;
        },
      );
    }
  }
}
