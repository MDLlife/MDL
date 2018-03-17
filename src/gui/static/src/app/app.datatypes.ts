/**
 * Internal Objects
 */

export class Address {
  address: string;
  coins: number;
  hours: number;
}

export class PurchaseOrder {
  coin_type: string;
  filename: string;
  deposit_address: string;
  recipient_address: string;
  status?: string;
}

export class TellerConfig {
  enabled: boolean;
  supported: string[];
  mdl_btc_exchange_rate: number;
  mdl_eth_exchange_rate: number;
  mdl_sky_exchange_rate: number;
  mdl_waves_exchange_rate: number;
  max_bound_addrs: number;

}

export class Transaction {
  addresses: string[];
  balance: number;
  block: number;
  confirmed: boolean;
  inputs: any[];
  outputs: any[];
  timestamp: number;
  txid: string;
}

export class Wallet {
  label: string;
  filename: string;
  seed: string;
  coins: number;
  hours: number;
  addresses: Address[];
  visible?: boolean;
  hideEmpty?: boolean;
  opened?: boolean;
}

/**
 * Response Objects
 */

export class GetWalletsResponseWallet {
  meta: GetWalletsResponseMeta;
  entries: GetWalletsResponseEntry[];
}

export class PostWalletNewAddressResponse {
  addresses: string[];
}

/**
 * Response Embedded Objects
 */

export class GetWalletsResponseMeta {
  label: string;
  filename: string;
  seed: string;
}

export class GetWalletsResponseEntry {
  address: string;
}
