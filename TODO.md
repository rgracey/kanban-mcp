- Add History / Audit Trail for Tickets
  Track status transitions, timestamps, edits, and comments.  
  **Why:** This is foundational data. Time tracking, gates, analytics, and accountability all depend on reliable event history. Without this, later features become hacks.

  Shift to an event-based model for tickets:

  ```
  TicketEvent {
    id
    ticket_id
    type (created, moved, assigned, checklist_toggled, etc.)
    actor
    timestamp
    payload (json)
  }
  ```

- Add Assignees
  Allow tickets to be owned by agents.

- Allow Ordering of Tickets by Priority
  Allow ordering of tickets by priority in MCP response (if it doesn't exist)

- Add Tasks (Checklists) Inside Tickets
  Checkable subtasks per ticket.

- Support URL Navigation to Specific Boards / Tickets (Move to Slugs)
  Deep-linkable state. Boards have a slug so URLs are easier to type (maybe also move tickets to slugs or something as well?)

- Time Tracking (Derived From History)
  Time in status, lead time, cycle time.

- Markdown Rendering for Descriptions (FE)
  Quality-of-life improvement.

- UI Overhaul
  - Strong identity -> clean and modern
  - Colour by status
  - Dark mode
  - Better editing flows
  - Explicit “Save changes” affordance in "edit ticket" UI

- Add SSE / WebSocket Reactivity for Frontend
  Live updates when board changes.

- General Code Review / Refactor

- README for Public Release
  Include the what/why, usage, dev setup.
