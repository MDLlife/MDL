import { Component, OnInit } from '@angular/core';
import 'rxjs/add/operator/takeWhile';
import { MatDialog } from '@angular/material';

import { AppService } from './services/app.service';
import { WalletService } from './services/wallet.service';
import { HwWalletService } from './services/hw-wallet.service';
import { HwPinDialogComponent } from './components/layout/hardware-wallet/hw-pin-dialog/hw-pin-dialog.component';
import { HwSeedWordDialogComponent } from './components/layout/hardware-wallet/hw-seed-word-dialog/hw-seed-word-dialog.component';
import { Bip39WordListService } from './services/bip39-word-list.service';
import { HwConfirmTxDialogComponent } from './components/layout/hardware-wallet/hw-confirm-tx-dialog/hw-confirm-tx-dialog.component';
import { LanguageService } from './services/language.service';
import { openChangeLanguageModal } from './utils';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
})
export class AppComponent implements OnInit {
  constructor(
    private appService: AppService,
    private languageService: LanguageService,
    walletService: WalletService,
    hwWalletService: HwWalletService,
    private bip38WordList: Bip39WordListService,
    private dialog: MatDialog,
  ) {
    hwWalletService.requestPinComponent = HwPinDialogComponent;
    hwWalletService.requestWordComponent = HwSeedWordDialogComponent;
    hwWalletService.signTransactionConfirmationComponent = HwConfirmTxDialogComponent;

    walletService.initialLoadFailed.subscribe(failed => {
      if (failed) {
        // The "?2" part indicates that error number 2 should be displayed.
        window.location.assign('assets/error-alert/index.html?2');
      }
    });
  }

  ngOnInit() {
    this.appService.testBackend();
    this.languageService.loadLanguageSettings();

    const subscription = this.languageService.selectedLanguageLoaded.subscribe(selectedLanguageLoaded => {
      if (!selectedLanguageLoaded) {
        openChangeLanguageModal(this.dialog, true).subscribe(response => {
          if (response) {
            this.languageService.changeLanguage(response);
          }
        });
      }

      subscription.unsubscribe();
    });
  }
}
