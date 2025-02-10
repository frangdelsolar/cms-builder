-- type SystemData struct {
-- 	gorm.Model
-- 	CreatedByID uint  `gorm:"not null" json:"createdById" jsonschema:"title=Created By Id,description=Id of the user who created this record"`
-- 	CreatedBy   *User `gorm:"foreignKey:CreatedByID" json:"createdBy" jsonschema:"title=Created By,description=User who created this record"`
-- 	UpdatedByID uint  `gorm:"not null" json:"updatedById" jsonschema:"title=Updated By Id,description=Id of the user who updated this record"`
-- 	UpdatedBy   *User `gorm:"foreignKey:UpdatedByID" json:"updatedBy" jsonschema:"title=Updated By,description=User who updated this record"`
-- }

-- type FileData struct {
-- 	*SystemData
-- 	Name string `json:"name"`
-- 	Path string `json:"path"` // relative path
-- 	Url  string `json:"url"`  // absolute path
-- }

-- type Upload struct {
-- 	*SystemData
-- 	*FileData
-- }

CREATE TABLE file_data (
    id INTEGER PRIMARY KEY,  
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by_id INTEGER NOT NULL,
    updated_by_id INTEGER NOT NULL,
    name TEXT,
    path TEXT,
    url TEXT,
    FOREIGN KEY (created_by_id) REFERENCES users(id),  
    FOREIGN KEY (updated_by_id) REFERENCES users(id)
);

CREATE TABLE uploads (
    id INTEGER PRIMARY KEY,  
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by_id INTEGER NOT NULL,
    updated_by_id INTEGER NOT NULL,
    file_data_id INTEGER NOT NULL,  
    FOREIGN KEY (created_by_id) REFERENCES users(id),
    FOREIGN KEY (updated_by_id) REFERENCES users(id),
    FOREIGN KEY (file_data_id) REFERENCES file_data(id)
);

-- Indexes for performance (optional but recommended)
CREATE INDEX idx_file_data_created_by_id ON file_data (created_by_id);
CREATE INDEX idx_file_data_updated_by_id ON file_data (updated_by_id);
CREATE INDEX idx_uploads_created_by_id ON uploads (created_by_id);
CREATE INDEX idx_uploads_updated_by_id ON uploads (updated_by_id);
CREATE INDEX idx_uploads_file_data_id ON uploads (file_data_id);
