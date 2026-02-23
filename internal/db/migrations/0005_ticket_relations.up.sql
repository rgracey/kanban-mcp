CREATE TABLE ticket_relations (
    from_ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    to_ticket_id   TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    kind           TEXT NOT NULL DEFAULT 'blocks',
    created_at     TEXT NOT NULL,
    PRIMARY KEY (from_ticket_id, to_ticket_id, kind)
);

CREATE INDEX ticket_relations_from ON ticket_relations(from_ticket_id);
CREATE INDEX ticket_relations_to   ON ticket_relations(to_ticket_id);
