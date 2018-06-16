const PROXY_CONFIG = {
  "/api/*": {
    "target": "http://0.0.0.0:46430",
    "pathRewrite": {
      "^/api" : ""
    },
    "secure": false,
    "logLevel": "debug",
    "bypass": function (req) {
      req.headers["origin"] = 'http://0.0.0.0:46430';
      req.headers["host"] = '0.0.0.0:46430';
      req.headers["referer"] = 'http://0.0.0.0:46430';
    }
},
  "/teller/*": {
    "target": "http://0.0.0.0:7071",
    "pathRewrite": {
      "^/teller" : "api/"
    },
    "secure": true,
    "logLevel": "debug"
  }
};

module.exports = PROXY_CONFIG;
