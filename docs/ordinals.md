# Detection of Inscriptions

The Babylon Staking API provides optional endpoints to check whether a UTXO 
contains an inscription. This functionality helps staking applications avoid 
spending specific UTXOs that might hold inscriptions. The API connects to the 
[Ordinal Service](https://github.com/ordinals/ord) and the Unisat API. Since 
Unisat is a paid service with rate limits, the API first attempts to get the 
UTXO status through the Ordinals Service. If that fails, it contacts the Unisat 
API as a backup to handle Ordinals Service downtime.

**Note**: This is an approximate solution and should not be considered a 
foolproof method. There may be false positives or negatives. If you intend to 
use this service to detect inscriptions, please assume that the service may not 
return entirely accurate results and implement additional fail-safe mechanisms 
for inscription detection.

To enable the optional ordinal API endpoint, provide the `ordinal` 
and `unisat` configurations under `assets`.

## Ordinal Service Client

The Ordinal Service Client is the primary method for checking inscriptions on 
UTXOs. It connects directly to a running instance of the [Ordinal Service](https://github.com/ordinals/ord).

### Verification Process

1. The `verifyViaOrdinalService` function is called with a list of UTXOs.
2. It uses the `FetchUTXOInfos` method of the Ordinals client to get 
information about the UTXOs.
   - For each UTXO, the Ordinals endpoint `/output/{TXID}:{OUTPUT_ID}` is invoked, 
   where `TXID` is the transaction ID and `OUTPUT_ID` is the output index(i.e `vout`).
3. For each UTXO, it checks:
   - If the `Runes` field is not empty and not "{}".
   - If the `Inscriptions` field is not empty.
4. If either condition is true, the UTXO is marked as having an inscription.

### Latency

The exact latency depends on the hardware and network setup of the services. 
As a reference, you can expect approximately 300ms for a steady 100 requests per second (rps).

## Unisat Service Client

More information about Unisat's Ordinal/BRC-20/Runes related endpoints can be 
found at: [Unisat API Documentation](https://docs.unisat.io/).

In our service, we utilize the following endpoint:
- `/v1/indexer/address/{{address}}/inscription-utxo-data`

### How to Use It

1. Log in at [Unisat Developer](https://developer.unisat.io/account/login) 
(create an account if you don't have one).
2. Copy the `API-Key`.
3. Set the key as an environment variable named `UNISAT_TOKEN`.
4. Configure the values for `unisat.host`, `limit`, `timeout`, etc. Refer 
to `config-docker.yml`.
5. Ensure you also set up the `ordinals` configuration, as this is a dependency.
6. Call the POST endpoint `/v1/ordinals/verify-utxos` as shown in the example below.
7. The Unisat API calls will only be triggered if the Ordinal Service is not 
responding or returns errors.

## Example POST Request

POST /v1/ordinals/verify-utxos
```json
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
```

Response:

```json
{
    "data": [
        {
            "txid": "143c33b4ff4450a60648aec6b4d086639322cb093195226c641ae4f0ae33c3f5",
            "vout": 0,
            "inscription": true
        },
        {
            "txid": "be3877c8dedd716f026cc77ef3f04f940b40b064d1928247cff5bb08ef1ba58e",
            "vout": 1,
            "inscription": false
        },
        {
            "txid": "d7f65a37f59088b3b4e4bc119727daa0a0dd8435a645c49e6a665affc109539d",
            "vout": 0,
            "inscription": false
        }
    ]
}
```