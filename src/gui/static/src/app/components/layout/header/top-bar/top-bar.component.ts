import {Component} from '@angular/core';
import {WalletService} from "../../../../services/wallet.service";
import {ApiService} from "../../../../services/api.service";
import {AppConfig} from "../../../../app.config";
import {IntervalObservable} from "rxjs/observable/IntervalObservable";
import {BlockchainService} from "../../../../services/blockchain.service";

@Component({
    selector: 'app-top-bar',
    templateUrl: './top-bar.component.html',
    styleUrls: ['./top-bar.component.scss']
})
export class TopBarComponent {
    current: number;
    highest: number;
    otcEnabled: boolean;
    percentage: number;
    version: string;

    constructor(
        public walletService: WalletService,
        private apiService: ApiService,
        private blockchainService: BlockchainService
    ) {
        this.otcEnabled = AppConfig.otcEnabled;
    }

    ngOnInit() {
        this.setVersion();
        IntervalObservable
            .create(3000)
            .flatMap(() => this.blockchainService.progress())
            .takeWhile((response: any) => !response.current || response.current !== response.highest)
            .subscribe(response => {
                    this.highest = response.highest;
                    this.current = response.current;
                    this.percentage = this.current && this.highest ? (this.current / this.highest * 100) : 0;
                    console.log(response);
                }, error => console.log(error),
                () => this.completeLoading());
    }

    loading() {
        return !this.current || !this.highest || this.current != this.highest;
    }

    private completeLoading() {
        this.current = 999999999999;
        this.highest = 999999999999;
        this.walletService.refreshBalances();
    }

    private setVersion() {
        return this.apiService.get('version')
            .subscribe(output => this.version = output.version);
    }
}
