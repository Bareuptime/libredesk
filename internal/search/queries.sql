-- name: search-conversations-by-reference-number
SELECT
    conversations.created_at,
    conversations.uuid,
    conversations.reference_number,
    conversations.subject
FROM conversations
WHERE reference_number::text = $1;

-- name: search-conversations-by-contact-email
SELECT
    conversations.created_at,
    conversations.uuid,
    conversations.reference_number,
    conversations.subject
FROM conversations
JOIN users ON conversations.contact_id = users.id
WHERE users.email = $1
ORDER BY conversations.created_at DESC
LIMIT 1000;

-- name: search-messages
SELECT
    c.created_at as "conversation_created_at",
    c.reference_number as "conversation_reference_number",
    c.uuid as "conversation_uuid",
    m.text_content
FROM conversation_messages m
    JOIN conversations c ON m.conversation_id = c.id
WHERE m.type != 'activity' and m.text_content ILIKE '%' || $1 || '%'
LIMIT 30;

-- name: search-contacts
SELECT 
    id,
    created_at,
    first_name,
    last_name,
    email
FROM users
WHERE type = 'contact'
AND deleted_at IS NULL
AND email ILIKE '%' || $1 || '%'
LIMIT 15;
