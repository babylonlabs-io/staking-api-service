# DB Design

## Logical Sharding in MongoDB

### Overview

**WARNING** If you plan to change the `logical-shard-count` to a higher number, 
consult the team first!

Logical sharding distributes data across multiple logical partitions or shards 
to enhance performance and scalability by reducing write contention. 
This method is especially beneficial in environments with high write throughput, 
where many operations update the same data points concurrently.

### Implementation

Sharding is achieved by randomly select a shard number to represent the document's ID
(or by appending a shard number to each document's ID, such as `{{docId}}:{{shardNumber}}`)

This approach effectively spreads 
writes across multiple documents, reducing bottlenecks.

#### Example
```go
func (db *Database) generateOverallStatsId() string {
	return fmt.Sprint(rand.Intn(int(db.cfg.LogicalShardCount)))
}
```

or 
```go
func (db *Database) generateXXXStatsId(docId string) string {
    randomShardNum := uint64(rand.Intn(int(db.cfg.LogicalShardCount)))
    return fmt.Sprintf("%s:%d", docId, randomShardNum)
}
```

### Considerations

#### Query Complexity

Logical sharding increases query complexity. 
Operations like `GetOverallStats` must now access multiple shards, 
leading to costlier queries as shard count increases.

#### Configuration Sensitivity

Increasing the LogicalShardCount can further complicate queries. 
Crucially, once increased, the shard count should never be decreased to avoid 
disrupting the sharding logic and causing data inconsistencies.

## Stats Locking

### Overview
Stats Locking refers to the use of a dedicated collection, `stats_lock`, in our service. 
This collection helps manage state and prevent duplication in statistical calculations, 
which is critical in a system designed to handle potential duplicate event messages.

### Implementation

The primary key for each document in the `stats_lock` collection is 
formatted as {{staking-tx-hash-hex}}:{{state}}, 
where we currently support two states: active and unbonded. Each document tracks 
whether specific calculations have been performed, 
preventing repetitive operations in case of message duplication.

This mechanism ensures that each stats calculation, 
whether adding (+) or subtracting (-), is performed only once per transaction, 
leveraging MongoDB transactions for consistency and reliability.

### Future Extension

To accommodate additional calculations in the future, 
the `stats_lock` collection may need to be extended with new boolean fields 
corresponding to each new calculation type. 
This approach will maintain the integrity and scalability of the stats locking system.

### Example

Consider `xyz` as the staking transaction hash hex.

1. **Service Reception**:
    - The service receives an `ActiveStakingEvent` which should add values to 
    the Total Value Locked (TVL).
  
2. **Breakdown of Stats Calculation**:
    - When processing overall stats (e.g., total TVL), our `stats_lock` collection 
    will record item with primary key of `xyz:active` with `overall_stats=true`, `finality_provider=false`.
    - When processing finality provider related stats (e.g., per finality provider TVL), 
    our `stats_lock` collection will update the record with primary key of  `xyz:active` 
    with `overall_stats=true`, `finality_provider=true`.
  
3. **Handling Duplicates**:
    - If our system receives the same `ActiveStakingEvent` again, 
    the stats calculation won't be reprocessed. 
    This is because the system checks the boolean values for `overall_stats` and 
    `finality_provider` individually, ensuring that each calculation is performed only once.