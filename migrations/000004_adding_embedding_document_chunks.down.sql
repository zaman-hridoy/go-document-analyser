

DROP INDEX IF EXISTS idx_document_chunks_embedding;

ALTER TABLE document_chunks
DROP COLUMN IF EXISTS embedding;