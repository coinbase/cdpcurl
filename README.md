# cdpcurl

`cdpcurl` is a tool that allows you to make HTTP requests to the Coinbase API with your CDP (Coinbase Developer Platform) API key. It is a wrapper around curl that automatically adds the necessary headers to authenticate your requests.

## Installation

## Homebrew

```bash
brew tap coinbase/cdpcurl
brew install cdpcurl
```

### AUR

```bash
yay -S cdpcurl-git
```

### Go

```bash
go install github.com/coinbase/cdpcurl@latest
```

## Example Usage

### Get account balance of BTC with Sign In With Coinbase API
```bash
cdpcurl -k ~/Downloads/cdp_api_key.json 'https://api.coinbase.com/v2/accounts/BTC'
```

### Get the latest price of BTC with Advanced Trading API
```bash
cdpcurl -k ~/Downloads/cdp_api_key.json 'https://api.coinbase.com/api/v3/brokerage/products/BTC-USDC'
```

### Create a wallet on Base Sepolia with Platform API

```bash
cdpcurl -k ~/Downloads/cdp_api_key.json -X POST -d '{"wallet": {"network_id": "base-sepolia"}}' 'https://api.developer.coinbase.com/platform/v1/wallets'
```
