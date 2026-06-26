# Human‑In‑The‑Loop Approval Workflow

## Overview
The **high‑value approval workflow** ensures that deals exceeding predefined risk/value thresholds are reviewed by a human before the autonomous sales engine can advance them.

## Trigger Conditions
- **Market‑cap tier**: `Enterprise`
- **Quoted pricing**: ≥ $100 000

If either condition is met, the deal state is set to **`Pending_Approval`** instead of automatically moving to `Negotiating` or `Closed_Won`.

## Technical Details
1. **State addition** – `StatePendingApproval` added to `internal/db/models.go`.
2. **Detection** – Helper `isHighValueDeal` in `internal/communication/engine.go` checks the two criteria.
3. **Engine change** – `LearningSalesEngine.Decide` now routes qualifying deals to `StatePendingApproval` and logs a message.
4. **Human action** – `Manager.ApproveDeal` method updates the deal state from `Pending_Approval` → `Negotiating` (callable via UI or API).
5. **A/B testing integration** – The approval step works seamlessly with existing template/objection success‑rate tracking.

## Future Integration
- **API endpoints** – `/api/deals/:id/approve` & `/api/deals/:id/reject`.
- **Notification** – Email/Slack alerts when a deal enters pending approval.
- **Metrics** – Track time‑to‑approval for operational insight.

## Impact
- Prevents autonomous progression of deals that could have high financial exposure.
- Provides a clear audit trail via state changes and logs.
- Keeps the existing autonomous flow untouched for low‑value opportunities.
