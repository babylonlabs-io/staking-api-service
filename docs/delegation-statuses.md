# Delegation Status Guide

This document explains all possible delegation statuses returned by the Staking API Service and what they mean for stakers.

## Overview

When you query the Staking API for delegation information, each delegation has a `status` field that indicates its current state in the staking lifecycle. There are **15 possible status values** that provide detailed information about where your delegation stands and what actions you can take.

## Status Categories

Statuses are organized into logical categories based on the delegation lifecycle:

1. **Setup States**: PENDING, VERIFIED, ACTIVE
2. **Unbonding States**: TIMELOCK_UNBONDING, EARLY_UNBONDING
3. **Withdrawable States**: TIMELOCK_WITHDRAWABLE, EARLY_UNBONDING_WITHDRAWABLE, TIMELOCK_SLASHING_WITHDRAWABLE, EARLY_UNBONDING_SLASHING_WITHDRAWABLE
4. **Withdrawn States**: TIMELOCK_WITHDRAWN, EARLY_UNBONDING_WITHDRAWN, TIMELOCK_SLASHING_WITHDRAWN, EARLY_UNBONDING_SLASHING_WITHDRAWN
5. **Special States**: SLASHED, EXPANDED

## State Architecture

Understanding how delegation states work across different layers of the Babylon ecosystem:

### Babylon Protocol Layer

The Babylon blockchain tracks delegations using 5 core states:
- **PENDING**: Awaiting covenant signatures
- **VERIFIED**: Has covenant quorum but no BTC inclusion proof
- **ACTIVE**: Has quorum + inclusion proof + within active period
- **UNBONDED**: Either unbonded early or timelock conditions met
- **EXPIRED**: Natural expiration occurred

**Code Reference**: `babylon/x/btcstaking/types/btcstaking.pb.go:37-66`

### Indexer Layer

The Staking Indexer enhances Babylon's states with:
- **Granular tracking**: Distinguishes between UNBONDING (in progress) vs UNBONDED (Babylon's term)
- **Sub-states**: Tracks the specific path (TIMELOCK vs EARLY_UNBONDING) and slashing variants
- **Additional states**: WITHDRAWABLE, WITHDRAWN, SLASHED, EXPANDED
- **State + SubState** model: Two fields that work together

**Example**: Indexer stores `State=UNBONDING, SubState=EARLY_UNBONDING`

**Documentation**: [Indexer State Definitions](https://github.com/babylonlabs-io/babylon-staking-indexer/blob/main/docs/states/overview.md)

### API Service Layer (This Document)

The Staking API **flattens** the indexer's state model into 15 user-friendly statuses:
- Combines State + SubState into a single status field
- Example: `State=UNBONDING, SubState=EARLY_UNBONDING` → `status=EARLY_UNBONDING`

**Code Reference**: `staking-api-service/internal/v2/types/delegation_states.go:69` (MapDelegationState function)

---

## Setup States

These are the initial states when a delegation is being set up.

### PENDING

**What it means**: Your delegation has been created on Babylon but is waiting for covenant committee signatures.

**What you can do**:
- Wait for covenant signatures to be collected
- Monitor the status until it transitions to VERIFIED or ACTIVE

**Next status**: VERIFIED (pre-approval flow) or ACTIVE (old flow)

**Typical duration**: On average, staking transactions are fully confirmed by the Bitcoin chain in 5 hours, depending on Bitcoin block times. Longer delays might be due to slow Bitcoin block times or using low network fees.

**Note**: The frontend may also show this as "Pending Verification" during the initial submission phase.

---

### VERIFIED

**What it means**: The covenant committee has provided the required signatures, but the system is waiting for confirmation that your staking transaction has been included in a Bitcoin block (with sufficient confirmations).

**What you can do**:
- Wait for the vigilante to report the inclusion proof to Babylon
- Your Bitcoin transaction needs to reach the required confirmation depth (typically 10 confirmations)

**Next status**: ACTIVE

**Typical duration**: Minutes to hours, depending on Bitcoin confirmations and vigilante reporting

**Note**: The frontend may show this as "Pending BTC Confirmation" while waiting for Bitcoin confirmations.

---

### ACTIVE

**What it means**: Your delegation is **fully active** and participating in the Babylon staking protocol. Your Bitcoin is contributing to the security of Babylon finality providers.

**What you can do**:
- Earn staking rewards (if applicable)
- Wait for the staking period to naturally expire, or
- Initiate early unbonding if you want to unlock your Bitcoin before the timelock expires

**Next status**:
- TIMELOCK_UNBONDING (if you wait for natural expiration)
- EARLY_UNBONDING (if you request early unbonding)
- SLASHED (if your finality provider misbehaves)
- EXPANDED (if you expand your delegation)

**Important**: This is the only state where your delegation is actively staking.

---

## Unbonding States

These states indicate your delegation is in the unbonding period and no longer actively staking.

### TIMELOCK_UNBONDING

**What it means**: Your delegation has reached its natural expiration and is in the unbonding period. The staking timelock is expiring, and the delegation is transitioning out of the active state.

**What you can do**:
- Wait for the unbonding period to complete
- Monitor when it becomes withdrawable

**Next status**: TIMELOCK_WITHDRAWABLE (when unbonding period completes)

**How you got here**: Your delegation automatically entered unbonding when it reached `endHeight - unbondingTime` blocks

**Typical duration**: Depends on the unbonding period configured in the staking parameters (typically measured in Bitcoin blocks)

**Note**: The delegation no longer contributes voting power to the finality provider during unbonding.

---

### EARLY_UNBONDING

**What it means**: You requested to unbond your delegation early (before the staking timelock expired), and the unbonding transaction has been submitted to Bitcoin. Your delegation is now in the unbonding waiting period.

**What you can do**:
- Wait for the unbonding period to complete
- Monitor when it becomes withdrawable

**Next status**:
- EARLY_UNBONDING_WITHDRAWABLE (when unbonding period completes normally)
- EARLY_UNBONDING_SLASHING_WITHDRAWABLE (if slashed during unbonding)

**How you got here**: You initiated an early unbonding request

**Typical duration**: Depends on the unbonding period configured in the staking parameters

---

## Withdrawable States

These states indicate your delegation funds can now be withdrawn from Bitcoin.

### TIMELOCK_WITHDRAWABLE

**What it means**: Your delegation went through natural expiration, completed the unbonding period, and your Bitcoin is now **ready to be withdrawn**. The timelock has fully expired.

**What you can do**:
- **Withdraw your Bitcoin immediately** by submitting a transaction spending the staking output via the timelock path
- Do not delay - funds can still be slashed even after timelock expires if you don't withdraw

**Next status**: TIMELOCK_WITHDRAWN (after you withdraw)

**How you got here**: Natural expiration path → unbonding period completed

**⚠️ Important**: Withdraw promptly! Funds remaining in outputs can still be slashed.

**Note**: While your withdrawal transaction is pending Bitcoin confirmation, the frontend may show this as "Withdrawing".

---

### EARLY_UNBONDING_WITHDRAWABLE

**What it means**: You requested early unbonding, the unbonding period has completed, and your Bitcoin is now **ready to be withdrawn**.

**What you can do**:
- **Withdraw your Bitcoin immediately** by submitting a transaction spending the unbonding output via the timelock path
- Do not delay - funds can still be slashed even after unbonding timelock expires

**Next status**: EARLY_UNBONDING_WITHDRAWN (after you withdraw)

**How you got here**: Early unbonding request → unbonding period completed

**⚠️ Important**: Withdraw promptly! Funds remaining in outputs can still be slashed.

---

### TIMELOCK_SLASHING_WITHDRAWABLE

**What it means**: Your staking transaction was slashed (due to finality provider misbehavior), and after the slashing timelock expired, your **remaining funds** (if any) are now ready to be withdrawn.

**What you can do**:
- **Withdraw any remaining Bitcoin** by submitting a transaction spending the slashing output
- Note that you may have lost some funds due to slashing penalties

**Next status**: TIMELOCK_SLASHING_WITHDRAWN (after you withdraw)

**How you got here**: Finality provider was slashed while your delegation was ACTIVE → slashing timelock expired

**⚠️ Note**: You likely lost some funds to slashing. Withdraw remaining funds promptly.

---

### EARLY_UNBONDING_SLASHING_WITHDRAWABLE

**What it means**: Your unbonding transaction was slashed (due to finality provider misbehavior during unbonding), and after the slashing timelock expired, your **remaining funds** (if any) are now ready to be withdrawn.

**What you can do**:
- **Withdraw any remaining Bitcoin** by submitting a transaction spending the slashing output
- Note that you may have lost some funds due to slashing penalties

**Next status**: EARLY_UNBONDING_SLASHING_WITHDRAWN (after you withdraw)

**How you got here**: You initiated early unbonding → finality provider was slashed during unbonding → slashing timelock expired

**⚠️ Note**: You likely lost some funds to slashing. Withdraw remaining funds promptly.

---

## Withdrawn States

These are terminal states indicating your delegation has been fully withdrawn.

### TIMELOCK_WITHDRAWN

**What it means**: You successfully withdrew your Bitcoin after going through the natural expiration and unbonding process. This is a **final state** - no further actions are possible.

**What you can do**: Nothing - your delegation lifecycle is complete. Your Bitcoin has been returned to you.

**How you got here**: Natural expiration path → unbonding completed → you withdrew funds

**Status**: ✅ Complete

---

### EARLY_UNBONDING_WITHDRAWN

**What it means**: You successfully withdrew your Bitcoin after requesting early unbonding and completing the unbonding period. This is a **final state** - no further actions are possible.

**What you can do**: Nothing - your delegation lifecycle is complete. Your Bitcoin has been returned to you.

**How you got here**: Early unbonding request → unbonding completed → you withdrew funds

**Status**: ✅ Complete

---

### TIMELOCK_SLASHING_WITHDRAWN

**What it means**: Your staking output was slashed, and you have withdrawn whatever funds remained after slashing penalties. This is a **final state** - no further actions are possible.

**What you can do**: Nothing - your delegation lifecycle is complete. You received whatever Bitcoin remained after slashing.

**How you got here**: Finality provider slashed while ACTIVE → slashing timelock expired → you withdrew remaining funds

**Status**: ⚠️ Complete (but slashed - you may have lost funds)

---

### EARLY_UNBONDING_SLASHING_WITHDRAWN

**What it means**: Your unbonding output was slashed during early unbonding, and you have withdrawn whatever funds remained after slashing penalties. This is a **final state** - no further actions are possible.

**What you can do**: Nothing - your delegation lifecycle is complete. You received whatever Bitcoin remained after slashing.

**How you got here**: Early unbonding request → finality provider slashed during unbonding → slashing timelock expired → you withdrew remaining funds

**Status**: ⚠️ Complete (but slashed - you may have lost funds)

---

## Special States

### SLASHED

**What it means**: Your delegation has been slashed because the finality provider you delegated to misbehaved (e.g., double-signed). Your funds are subject to slashing penalties.

**What you can do**:
- Wait for the slashing timelock to expire
- After expiry, withdraw any remaining funds

**Next status**:
- TIMELOCK_SLASHING_WITHDRAWABLE (when slashing timelock expires)
- EARLY_UNBONDING_SLASHING_WITHDRAWABLE (if slashed during unbonding)

**How you got here**: Your finality provider committed a slashable offense

**⚠️ Warning**: You may lose a portion of your staked Bitcoin due to slashing penalties. This emphasizes the importance of choosing reputable finality providers.

---

### EXPANDED

**What it means**: Your delegation has been expanded/extended into a new delegation (by spending the staking output as an input to a new staking transaction with an extended timelock). This is a **final state** for the original delegation.

**What you can do**:
- Check for the new delegation with extended timelock
- The original delegation is now complete and has been replaced

**How you got here**: You submitted an expansion transaction to extend the staking period before the original delegation's timelock expired, allowing you to keep staking without going through the full restaking process

**Status**: ✅ Complete (expanded/extended into new delegation)

**Note**: Stake expansion currently supports **extending the time** (timelock), not increasing the stake amount. This allows you to keep your delegation active without unbonding and restaking.

---

## Common User Journeys

### Happy Path (Natural Expiration)

```
PENDING → VERIFIED → ACTIVE → TIMELOCK_UNBONDING → TIMELOCK_WITHDRAWABLE → TIMELOCK_WITHDRAWN
```

**Timeline**:
1. Create delegation (PENDING)
2. Covenant signs (VERIFIED)
3. Stake actively participates (ACTIVE)
4. Timelock expires naturally (TIMELOCK_UNBONDING)
5. Unbonding completes (TIMELOCK_WITHDRAWABLE)
6. You withdraw (TIMELOCK_WITHDRAWN)

---

### Early Unbonding Path

```
PENDING → VERIFIED → ACTIVE → EARLY_UNBONDING → EARLY_UNBONDING_WITHDRAWABLE → EARLY_UNBONDING_WITHDRAWN
```

**Timeline**:
1. Create delegation (PENDING)
2. Covenant signs (VERIFIED)
3. Stake actively participates (ACTIVE)
4. **You request early unbonding** (EARLY_UNBONDING)
5. Unbonding period completes (EARLY_UNBONDING_WITHDRAWABLE)
6. You withdraw (EARLY_UNBONDING_WITHDRAWN)

---

### Slashing During Active Staking

```
PENDING → VERIFIED → ACTIVE → SLASHED → TIMELOCK_SLASHING_WITHDRAWABLE → TIMELOCK_SLASHING_WITHDRAWN
```

**Timeline**:
1. Create delegation (PENDING)
2. Covenant signs (VERIFIED)
3. Stake actively participates (ACTIVE)
4. **Finality provider misbehaves** (SLASHED)
5. Slashing timelock expires (TIMELOCK_SLASHING_WITHDRAWABLE)
6. You withdraw remaining funds (TIMELOCK_SLASHING_WITHDRAWN)

---

## API Integration

When integrating with the Staking API, you can query delegation status via:

```
GET /v2/delegations/{staking_tx_hash}
```

The response includes a `status` field with one of the 15 values documented above.

### Example Response

```json
{
  "staking_tx_hash": "abc123...",
  "status": "EARLY_UNBONDING_WITHDRAWABLE",
  "staking_amount": 100000000,
  "finality_provider": "fp_pubkey_hex",
  ...
}
```

### Recommended Status Handling

For user interfaces, we recommend grouping statuses into actionable categories:

**Active Staking**:
- `ACTIVE`

**Processing**:
- `PENDING`, `VERIFIED`

**Unbonding** (informational):
- `TIMELOCK_UNBONDING`, `EARLY_UNBONDING`

**Action Required - Withdraw Now**:
- `TIMELOCK_WITHDRAWABLE`
- `EARLY_UNBONDING_WITHDRAWABLE`
- `TIMELOCK_SLASHING_WITHDRAWABLE`
- `EARLY_UNBONDING_SLASHING_WITHDRAWABLE`

**Complete**:
- `TIMELOCK_WITHDRAWN`
- `EARLY_UNBONDING_WITHDRAWN`
- `TIMELOCK_SLASHING_WITHDRAWN`
- `EARLY_UNBONDING_SLASHING_WITHDRAWN`
- `EXPANDED`

**Problem - Slashed**:
- `SLASHED`

---

## Further Reading

- [Indexer State Definitions](https://github.com/babylonlabs-io/babylon-staking-indexer/blob/main/docs/states/overview.md)
- [Indexer State Lifecycle](https://github.com/babylonlabs-io/babylon-staking-indexer/blob/main/docs/states/lifecycle.md)
- [Staking API Service Architecture](../README.md)
