export const AppConfig = {
  otcEnabled: true,
  maxHardwareWalletAddresses: 1,
  useHwWalletDaemon: true,
  urlForHwWalletVersionChecking: 'https://version.skycoin.com/skywallet/version.txt',
  hwWalletDownloadUrlAndPrefix: 'https://downloads.skycoin.com/skywallet/skywallet-firmware-v',

  urlForVersionChecking: 'https://version.skycoin.com/skycoin/version.txt',
  walletDownloadUrl: 'https://www.skycoin.com/downloads/',

  /**
   * This wallet uses the Skycoin URI Specification (based on BIP-21) when creating QR codes and
   * requesting coins. This variable defines the prefix that will be used for creating QR codes
   * and URLs. IT MUST BE UNIQUE FOR EACH COIN.
   */
  uriSpecificatioPrefix: 'mdl',

  languages: [{
      code: 'en',
      name: 'English',
      iconName: 'en.png',
    },
    {
      code: 'zh',
      name: '中文',
      iconName: 'zh.png',
    },
    {
      code: 'es',
      name: 'Español',
      iconName: 'es.png',
    },
  ],
  defaultLanguage: 'en',
};
