CREATE TABLE IF NOT EXISTS roles_permissions (
    role_id bigint REFERENCES roles(id) ON DELETE CASCADE,
    permission_id bigint REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
