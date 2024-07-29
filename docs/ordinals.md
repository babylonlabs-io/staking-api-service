# Ordinals in API system

The Babylon Staking API allows for the option to deploy additional endpoints  
that check whether a UTXO contains an inscription or not, with the aim to help  
staking applications identify whether they should avoid spending a particular UTXO.  
This is accomplished through a connection to the  
[Ordinal Service](https://github.com/ordinals/ord) and  
a connection to the Unisat API.  
Due to Unisat being a payed service and applying rate limits,  
the API initially tries to get the status of a UTXO through the Ordinals Service,  
and if that fails, then contacts the Unisat API,  
effectively using it as a back-up mechanism to handle downtime from the Ordinals Service.  

NOTE: To enable the optional ordinal API endpoint, you will need to provide the 
`ordinal` and `unisat` configurations under `assets`

## Ordinal service Client

The Ordinal Service Client is the primary method for checking inscriptions on UTXOs. It connects directly to a running instance of the [Ordinal Service](https://github.com/ordinals/ord).

### Verification Process

1. The `verifyViaOrdinalService` function is called with a list of UTXOs.
2. It uses the `FetchUTXOInfos` method of the Ordinals client to get information
   about the UTXOs.
   - Underlyingly, for every UTXO the Ordinals endpoint `/output/TXID:OUTPUT_ID`
     is invoked, where `TXID` the transaction hex and `OUTPUT_ID` the output of
     interest
3. For each UTXO, it checks:
   - If the `Runes` field is not empty and not "{}"
   - If the `Inscriptions` field is not empty
4. If either condition is true, the UTXO is marked as having an inscription.

## Unisat Service Client

You can find more information about Unisat's Ordinal/BRC-20/Runes related endpoints at:
https://docs.unisat.io/

In our service, we only utilize the following endpoint:
- `/v1/indexer/address/{{address}}/inscription-utxo-data`

### How to Use It

1. Log in via https://developer.unisat.io/account/login (create an account if you don't have one).
2. Copy the `API-Key`.
3. Set the key as an environment variable named `UNISAT_TOKEN`.
4. Configure the values for `unisat.host`, `limit`, `timeout`, etc. Refer to `config-docker.yml`.
5. Ensure you also set up the `ordinals` configuration, as this is a dependency.
6. Call the POST endpoint `/v1/ordinals/verify-utxos` as shown in the example below:
7. The calls to unisat will only be triggered if the ordinal service is not responding or returning errors
```POST
{
    "utxos": [
        {
            "txid": "143c33b4ff4450a60648aec6b4d086639322cb093195226c641ae4f0ae33c3f5",
            "vout": 2
        },
        {
            "txid": "be3877c8dedd716f026cc77ef3f04f940b40b064d1928247cff5bb08ef1ba58e",
            "vout": 0
        },
        {
            "txid": "d7f65a37f59088b3b4e4bc119727daa0a0dd8435a645c49e6a665affc109539d",
            "vout": 0
        }
    ],
    "address": "tb1pyqjxwcdv6pfcaj2l565ludclz2pwu2k5azs6uznz8kml74kkma6qm0gzlv"
}
