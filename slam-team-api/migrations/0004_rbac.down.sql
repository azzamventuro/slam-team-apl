-- Rollback of 0004_rbac. Drop in reverse dependency order.
-- Enums (aksi_permission, cakupan_permission) belong to 0001_enums and
-- roll back with that migration, not here.

DROP TABLE IF EXISTS user_role;
DROP TABLE IF EXISTS role_permission;
DROP TABLE IF EXISTS mst_role;
DROP TABLE IF EXISTS mst_permission;
DROP TABLE IF EXISTS mst_modul;
