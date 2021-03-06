package stmt

const CreateTables = `

CREATE TABLE IF NOT EXISTS file
(
  id            text    PRIMARY KEY,
  name          text    NOT NULL,
  count         int     NOT NULL,
  size          int     NOT NULL,
  type          text    NOT NULL,
  thumb         int     NOT NULL,
  hash          text    NOT NULL UNIQUE,
  like          int     NOT NULL,
  ctime         int     NOT NULL,
  utime         int     NOT NULL,
  checked       int     NOT NULL,
  damaged       int     NOT NULL,
  deleted       int     NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_file_name ON file(name);
CREATE INDEX IF NOT EXISTS idx_file_hash ON file(hash);
CREATE INDEX IF NOT EXISTS idx_file_ctime ON file(ctime);
CREATE INDEX IF NOT EXISTS idx_file_utime ON file(utime);
CREATE INDEX IF NOT EXISTS idx_file_checked ON file(checked);

CREATE TABLE IF NOT EXISTS tag
(
  id            text    PRIMARY KEY,
  ctime         int     NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tag_ctime ON tag(ctime);

CREATE TABLE IF NOT EXISTS file_tag
(
  file_id   text    REFERENCES file(id) ON DELETE CASCADE,
  tag_id    text    REFERENCES tag(id) ON UPDATE CASCADE ON DELETE CASCADE,
  UNIQUE (file_id, tag_id)
);

CREATE TABLE IF NOT EXISTS taggroup
(
  id            text    PRIMARY KEY,
  tags          blob    NOT NULL UNIQUE,
  protected     int     NOT NULL,
  utime         int     NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_taggroup_utime ON taggroup(utime);

CREATE TABLE IF NOT EXISTS metadata
(
  name         text    NOT NULL UNIQUE,
  int_value    int     NOT NULL DEFAULT 0,
  text_value   text    NOT NULL DEFAULT "" 
);
`

const InsertIntValue = `INSERT INTO metadata (name, int_value) VALUES (?, ?);`
const GetIntValue = `SELECT int_value FROM metadata WHERE name=?;`
const UpdateIntValue = `UPDATE metadata SET int_value=? WHERE name=?;`

const InsertTextValue = `INSERT INTO metadata (name, text_value) VALUES (?, ?);`
const GetTextValue = `SELECT text_value FROM metadata WHERE name=?;`
const UpdateTextValue = `UPDATE metadata SET text_value=? WHERE name=?;`

const GetFile = `SELECT * FROM file WHERE id=?;`
const GetFileName = `SELECT name FROM file WHERE id=?`
const GetFileHash = `SELECT hash FROM file WHERE id=?;`
const GetFileDamaged = `SELECT damaged FROM file WHERE id=?;`
const GetFileID = `SELECT id FROM file WHERE hash=?;`
const GetFileIDsByName = `SELECT id FROM file WHERE name=? ORDER BY ctime DESC;`
const CountFilesByName = `SELECT count(*) FROM file WHERE name=?;`
const SetFilesCount = `UPDATE file SET count=? WHERE name=?;`
const GetFiles = `SELECT * FROM file WHERE deleted=0 ORDER BY ctime DESC LIMIT ?;`
const GetImages = `
  SELECT * FROM file WHERE deleted=0 and type like "image/%"
  ORDER BY ctime DESC LIMIT ?;`
const GetDeletedFiles = `SELECT * FROM file WHERE deleted>0 ORDER BY utime;`
const GetAllFiles = `SELECT * FROM file;`
const GetFilesNeedCheck = `SELECT * FROM file WHERE checked<?;`
const CountFilesNeedCheck = `SELECT count(*) FROM file WHERE checked<?;`
const SetFileChecked = `UPDATE file SET checked=?, damaged=? WHERE id=?;`
const InsertFile = `INSERT INTO file (
  id, name, count, size, type, thumb, hash, like, ctime, utime, checked, damaged, deleted)
  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
const ReplaceFile = `UPDATE file SET size=?, hash=?, utime=?, checked=0, damaged=0 WHERE id=?;`
const SetFileDeletedNow = `UPDATE file SET deleted=?, utime=? WHERE id=?;`
const RenameFilesNow = `UPDATE file SET name=?, type=?, utime=? WHERE name=?;`
const UpdateNow = `UPDATE file SET utime=? WHERE id=?;`
const CountAllFiles = `SELECT count(*) FROM file;`
const CountDamagedFiles = `SELECT count(*) FROM file WHERE damaged>0;`
const DamagedFiles = `SELECT * FROM file WHERE damaged>0;`
const DamagedFileIDs = `SELECT id FROM file WHERE damaged>0;`
const DeleteFile = `DELETE FROM file WHERE id=?;`

const TotalSize = `SELECT COALESCE(sum(size),0) as totalsize FROM file;`

const GetTag = `SELECT * FROM tag WHERE id=?;`
const GetTagCTime = `SELECT ctime FROM tag WHERE id=?;`
const InsertTag = `INSERT INTO tag (id, ctime) VALUES ( ?, ?);`
const InsertFileTag = `INSERT INTO file_tag (file_id, tag_id) VALUES (?, ?);`
const DeleteTag = `DELETE FROM tag WHERE id=?;`
const DeleteTags = `DELETE FROM file_tag WHERE file_id=? and tag_id=?;`
const RenameTag = `UPDATE tag SET id=? WHERE id=?; `
const RenameTag_InFileTag = `UPDATE file_tag SET tag_id=? WHERE tag_id=?;`

const AllTagGroups = `SELECT * FROM taggroup ORDER BY utime;`
const GetTagGroupID = `SELECT id FROM taggroup WHERE tags=?;`
const InsertTagGroup = `INSERT INTO taggroup
    (id, tags, protected, utime) VALUES (?, ?, ?, ?);`
const UpdateTagGroupNow = `UPDATE taggroup SET utime=? WHERE id=?;`
const TagGroupCount = `SELECT count(*) FROM taggroup`
const LastTagGroup = `SELECT id FROM taggroup WHERE protected=0
    ORDER BY utime LIMIT 1;`
const DeleteTagGroup = `DELETE FROM taggroup WHERE id=?;`
const SetTagGroupProtected = `UPDATE taggroup SET protected=? WHERE id=?;`

const GetTagsByFile = `SELECT tag_id FROM file_tag WHERE file_id=?;`

const SearchFileName = `
  SELECT * FROM file WHERE deleted=0 and name LIKE ? ORDER BY ctime DESC;`
const SearchImageName = `
  SELECT * FROM file WHERE deleted=0 and type like "image/%" and name LIKE ?
  ORDER BY ctime DESC;`

const GetFilesByTag = `SELECT file.id FROM tag
    INNER JOIN file_tag ON tag.id = file_tag.tag_id
    INNER JOIN file ON file_tag.file_id = file.id
    WHERE file.deleted=0 and tag.id=?;`

const GetImagesByTag = `SELECT file.id FROM tag
    INNER JOIN file_tag ON tag.id = file_tag.tag_id
    INNER JOIN file ON file_tag.file_id = file.id
    WHERE file.deleted=0 and file.type like "image/%" and tag.id=?;`

const GetFilesHasThumbByTag = `SELECT file.id FROM tag
    INNER JOIN file_tag ON tag.id = file_tag.tag_id
    INNER JOIN file ON file_tag.file_id = file.id
    WHERE file.deleted=0 and file.thumb=1 and tag.id=?
    ORDER BY file.ctime DESC;`

const GetFilesNoThumbByTag = `SELECT file.id FROM tag
    INNER JOIN file_tag ON tag.id = file_tag.tag_id
    INNER JOIN file ON file_tag.file_id = file.id
    WHERE file.deleted=0 and file.thumb=0 and tag.id=?
    ORDER BY file.ctime DESC;`

const AllFilesByTag = `SELECT file_id FROM file_tag WHERE tag_id=?;`

const AllTagsByName = `SELECT tag.id, tag.ctime,
    (SELECT count(file.id) FROM file
     INNER JOIN file_tag ON file.id = file_tag.file_id
     WHERE file.deleted=0 and file_tag.tag_id = tag.id)
    FROM tag ORDER BY tag.id;`

const AllTagsByDate = `SELECT tag.id, tag.ctime,
    (SELECT count(*) FROM file
     INNER JOIN file_tag ON file.id = file_tag.file_id
     WHERE file.deleted=0 and file_tag.tag_id = tag.id)
    FROM tag ORDER BY tag.ctime;`
