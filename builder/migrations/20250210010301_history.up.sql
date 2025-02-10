-- type HistoryEntry struct {
-- 	gorm.Model
-- 	User         *User      `json:"user"`
-- 	UserId       string     `gorm:"foreignKey:UserId" json:"userId"`
-- 	Username     string     `json:"username"`
-- 	Action       CRUDAction `json:"action"`
-- 	ResourceName string     `json:"resourceName"`
-- 	ResourceId   string     `json:"resourceId"`
-- 	Timestamp    string     `gorm:"type:timestamp" json:"timestamp"`
-- 	Detail       string     `json:"detail"`
-- }

CREATE TABLE history_entries (
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE, 
    user_id TEXT REFERENCES users(id),  
    username TEXT,
    action TEXT,  
    resource_name TEXT,
    resource_id TEXT,
    timestamp TIMESTAMP,  
    detail TEXT
);

CREATE INDEX idx_history_entries_user_id ON history_entries (user_id);
CREATE INDEX idx_history_entries_resource_name ON history_entries (resource_name);
CREATE INDEX idx_history_entries_resource_id ON history_entries (resource_id);
CREATE INDEX idx_history_entries_timestamp ON history_entries (timestamp); -- Index the timestamp for querying by date/time ranges
