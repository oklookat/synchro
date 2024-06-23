PRAGMA foreign_keys = ON;

------ REMOTE
CREATE TABLE IF NOT EXISTS remote (
    id TEXT PRIMARY KEY,
    is_enabled INTEGER NOT NULL DEFAULT 1,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS account (
    id TEXT PRIMARY KEY,
    remote_name INTEGER NOT NULL REFERENCES remote (name) ON DELETE CASCADE,
    alias TEXT NOT NULL DEFAULT 'Without alias',
    auth TEXT NOT NULL,
    added_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS account_settings (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL UNIQUE REFERENCES account (id) ON DELETE CASCADE,
    sync_liked_albums INTEGER NOT NULL DEFAULT 0,
    sync_liked_artists INTEGER NOT NULL DEFAULT 0,
    sync_liked_tracks INTEGER NOT NULL DEFAULT 0,
    sync_playlists INTEGER NOT NULL DEFAULT 0,
    last_sync_liked_albums INTEGER NOT NULL DEFAULT 0,
    last_sync_liked_artists INTEGER NOT NULL DEFAULT 0,
    last_sync_liked_tracks INTEGER NOT NULL DEFAULT 0,
    last_sync_playlists INTEGER NOT NULL DEFAULT 0
);

------ SNAPSHOTS
CREATE TABLE IF NOT EXISTS snapshot (
    id TEXT PRIMARY KEY,
    remote_name INTEGER NOT NULL REFERENCES remote (name) ON DELETE CASCADE,
    alias TEXT DEFAULT 'Where an alias?',
    auto INTEGER NOT NULL DEFAULT 0,
    restoreable_liked_albums INTEGER NOT NULL DEFAULT 0,
    restoreable_liked_artists INTEGER NOT NULL DEFAULT 0,
    restoreable_liked_tracks INTEGER NOT NULL DEFAULT 0,
    restoreable_playlists INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS snapshot_liked_album (
    id TEXT PRIMARY KEY,
    snapshot_id TEXT NOT NULL REFERENCES snapshot (id) ON DELETE CASCADE,
    id_on_remote TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS snapshot_liked_artist (
    id TEXT PRIMARY KEY,
    snapshot_id TEXT NOT NULL REFERENCES snapshot (id) ON DELETE CASCADE,
    id_on_remote TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS snapshot_liked_track (
    id TEXT PRIMARY KEY,
    snapshot_id TEXT NOT NULL REFERENCES snapshot (id) ON DELETE CASCADE,
    id_on_remote TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS snapshot_playlist (
    id TEXT PRIMARY KEY,
    snapshot_id TEXT NOT NULL REFERENCES snapshot (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    is_visible INTEGER NOT NULL DEFAULT 0,
    description TEXT DEFAULT NULL,
    created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS snapshot_playlist_track (
    id TEXT PRIMARY KEY,
    snapshot_id TEXT NOT NULL REFERENCES snapshot_playlist (id) ON DELETE CASCADE,
    id_on_remote TEXT NOT NULL
);

------ ARTISTS
CREATE TABLE IF NOT EXISTS artist (id TEXT PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_artist (
    id TEXT PRIMARY KEY,
    entity_id INTEGER NOT NULL REFERENCES artist (id) ON DELETE CASCADE,
    remote_name TEXT NOT NULL REFERENCES remote (name) ON DELETE CASCADE,
    id_on_remote TEXT DEFAULT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, remote_name, id_on_remote)
);

CREATE TABLE IF NOT EXISTS synced_artist (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL UNIQUE REFERENCES artist (id) ON DELETE CASCADE,
    is_synced INTEGER NOT NULL DEFAULT 1,
    is_synced_modified_at INTEGER NOT NULL DEFAULT 0
);

------ ALBUMS
CREATE TABLE IF NOT EXISTS album (id TEXT PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_album (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL REFERENCES album (id) ON DELETE CASCADE,
    remote_name TEXT NOT NULL REFERENCES remote (name) ON DELETE CASCADE,
    id_on_remote TEXT DEFAULT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, remote_name, id_on_remote)
);

CREATE TABLE IF NOT EXISTS synced_album (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL UNIQUE REFERENCES album (id) ON DELETE CASCADE,
    is_synced INTEGER NOT NULL DEFAULT 1,
    is_synced_modified_at INTEGER NOT NULL DEFAULT 0
);

------ TRACKS
CREATE TABLE IF NOT EXISTS track (id TEXT PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_track (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL REFERENCES track (id) ON DELETE CASCADE,
    remote_name TEXT NOT NULL REFERENCES remote (name) ON DELETE CASCADE,
    id_on_remote TEXT DEFAULT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, remote_name, id_on_remote)
);

CREATE TABLE IF NOT EXISTS synced_track (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL UNIQUE REFERENCES track (id) ON DELETE CASCADE,
    is_synced INTEGER NOT NULL DEFAULT 1,
    is_synced_modified_at INTEGER NOT NULL DEFAULT 0
);

------ PLAYLISTS
CREATE TABLE IF NOT EXISTS playlist (id TEXT PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_playlist (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL REFERENCES playlist (id) ON DELETE CASCADE,
    remote_name TEXT NOT NULL REFERENCES account (id) ON DELETE CASCADE,
    id_on_remote TEXT NOT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, remote_name, id_on_remote)
);

CREATE TABLE IF NOT EXISTS synced_playlist (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL REFERENCES playlist (id) ON DELETE CASCADE,
    is_synced INTEGER NOT NULL DEFAULT 1,
    is_synced_modified_at INTEGER NOT NULL DEFAULT 0,
    is_visible INTEGER NOT NULL DEFAULT 0,
    is_visible_modified_at INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL DEFAULT 'synchrodummy',
    name_modified_at INTEGER NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    description_modified_at INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS synced_playlist_track (
    id TEXT PRIMARY KEY,
    synced_playlist_id TEXT NOT NULL REFERENCES synced_playlist (id) ON DELETE CASCADE,
    entity_id TEXT NOT NULL REFERENCES track (id) ON DELETE CASCADE,
    is_synced INTEGER NOT NULL DEFAULT 1,
    is_synced_modified_at INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS account_synced_playlist_settings (
    id TEXT PRIMARY KEY,
    playlist_id TEXT NOT NULL REFERENCES playlist (id) ON DELETE CASCADE,
    account_id TEXT NOT NULL REFERENCES account (id) ON DELETE CASCADE,
    sync_name INTEGER NOT NULL DEFAULT 1,
    sync_description INTEGER NOT NULL DEFAULT 1,
    sync_visibility INTEGER NOT NULL DEFAULT 1,
    sync_tracks INTEGER NOT NULL DEFAULT 1,
    last_sync_name INTEGER NOT NULL DEFAULT 0,
    last_sync_description INTEGER NOT NULL DEFAULT 0,
    last_sync_visibility INTEGER NOT NULL DEFAULT 0,
    last_sync_tracks INTEGER NOT NULL DEFAULT 0,
    UNIQUE (playlist_id, account_id)
);