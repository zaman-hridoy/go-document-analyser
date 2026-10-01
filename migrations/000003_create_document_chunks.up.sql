CREATE TABLE document_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    document_id UUID NOT NULL
        REFERENCES documents(id)
        ON DELETE CASCADE,

    chunk_index INT NOT NULL,
    content TEXT NOT NULL,

    page_number INT,
    token_count INT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(document_id, chunk_index)
);

CREATE INDEX idx_document_chunks_document_id
ON document_chunks(document_id);