CREATE TABLE chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL DEFAULT 'New chat',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE chat_documents (
    chat_id UUID NOT NULL
        REFERENCES chats(id)
        ON DELETE CASCADE,

    document_id UUID NOT NULL
        REFERENCES documents(id)
        ON DELETE CASCADE,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chat_id, document_id)
);

CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    chat_id UUID NOT NULL
        REFERENCES chats(id)
        ON DELETE CASCADE,
    
    role TEXT NOT NULL
        CHECK(role IN ('user', 'assistant')),
    
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE INDEX idx_chat_messages_chat_id_created_at
ON chat_messages(chat_id, created_at);

CREATE INDEX idx_chat_documents_document_id
ON chat_documents(document_id);

