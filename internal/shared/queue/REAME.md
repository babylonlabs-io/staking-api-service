# Event Message Processing Guidelines

In our system, message delivery adheres to three foundational principles to ensure robust and reliable event handling:

1. **Guarantee of Message Delivery**: No message will be lost.
2. **Handling Out-of-Order Messages**: Messages may arrive in an unexpected order.
3. **Duplicate Message Management**: The system can receive duplicate messages.

When designing or implementing handlers, adherence to these principles is paramount. This ensures that event processing is resilient, idempotent, and consistent.

## Implementation Example: ActiveStakingHandler

The `ActiveStakingHandler` exemplifies our approach to preserving these rules through careful design and strategic method ordering.

### Design Rationale

This handler prioritizes event processing integrity, especially in scenarios involving service interruptions. By positioning the `SaveActiveStakingDelegation` operation at the end, we mitigate risks associated with partial processing due to unexpected system restarts or failures.

### Problem Addressed

Interruptions after a state update but before completion of all intended operations could lead to requeued messages being incorrectly assumed as fully processed. This misassumption risks omitting unexecuted operations, compromising data integrity and processing completeness.

### Solution Strategy

To counteract this, we ensure that state-changing operations occur only after all checks and preparatory actions have succeeded. If a message is reprocessed, the system reattempts all operations, safeguarding against lost actions. This approach demands each component be capable of handling duplicates and out-of-order messages effectively.

## TL;DR: Event Processing Steps

1. **Eligibility Verification**: Initially, verify the event's relevance and timeliness. Disregard or requeue messages as appropriate based on their current applicability.
2. **Custom Logic Execution**: Implement your specific logic (e.g., statistics calculations, expiration checks) with resilience to duplication and out-of-order scenarios.
3. **State Alteration**: Conclude with state-changing actions, ensuring no prior steps are skipped or lost due to message reprocessing.

By following these guidelines, handlers within the system maintain a high degree of resilience and data integrity, even in the face of challenges like service interruptions and message anomalies.
