# Detection of Inscriptions

The Babylon Staking API provides optional endpoints to check whether a UTXO 
contains an inscription. This functionality helps staking applications avoid 
spending specific UTXOs that might hold inscriptions. The API connects to the 
[Ordinal Service](https://github.com/ordinals/ord) only.

**Note**: This is an approximate solution and should not be considered a 
foolproof method. There may be false positives or negatives. If you intend to 
use this service to detect inscriptions, please assume that the service may not 
return entirely accurate results and implement additional fail-safe mechanisms 
for inscription detection.

To enable the optional ordinal API endpoint, provide the `ordinal` configurations
under `assets`.

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