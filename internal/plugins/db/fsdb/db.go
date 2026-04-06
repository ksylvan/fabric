package fsdb

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/danielmiessler/fabric/internal/i18n"
	"github.com/joho/godotenv"
)

const (
	ConfigDirPerms os.FileMode = 0o700
	EnvFilePerms   os.FileMode = 0o600
)

func NewDb(dir string) (db *Db) {

	db = &Db{Dir: dir}

	db.EnvFilePath = db.FilePath(".env")

	db.Patterns = &PatternsEntity{
		StorageEntity:          &StorageEntity{Label: "Patterns", Dir: db.FilePath("patterns"), ItemIsDir: true},
		SystemPatternFile:      "system.md",
		UniquePatternsFilePath: db.FilePath("unique_patterns.txt"),
		CustomPatternsDir:      "", // Will be set after loading .env file
	}

	db.Sessions = &SessionsEntity{
		&StorageEntity{Label: "Sessions", Dir: db.FilePath("sessions"), FileExtension: ".json"}}

	db.Contexts = &ContextsEntity{
		&StorageEntity{Label: "Contexts", Dir: db.FilePath("contexts")}}

	return
}

type Db struct {
	Dir string

	Patterns *PatternsEntity
	Sessions *SessionsEntity
	Contexts *ContextsEntity

	EnvFilePath string
}

func (o *Db) Configure() (err error) {
	if err = os.MkdirAll(o.Dir, ConfigDirPerms); err != nil {
		return
	}

	if err = o.ensureSecureConfigDirPerms(); err != nil {
		return
	}

	if err = o.LoadEnvFile(); err != nil {
		return
	}

	// Set custom patterns directory after loading .env file
	customPatternsDir := os.Getenv("CUSTOM_PATTERNS_DIRECTORY")
	if customPatternsDir != "" {
		// Expand home directory if needed
		if strings.HasPrefix(customPatternsDir, "~/") {
			if homeDir, err := os.UserHomeDir(); err == nil {
				customPatternsDir = filepath.Join(homeDir, customPatternsDir[2:])
			}
		}
		o.Patterns.CustomPatternsDir = customPatternsDir
	}

	if err = o.Patterns.Configure(); err != nil {
		return
	}

	if err = o.Sessions.Configure(); err != nil {
		return
	}

	if err = o.Contexts.Configure(); err != nil {
		return
	}

	return
}

func (o *Db) LoadEnvFile() (err error) {
	if err = o.ensureSecureEnvFilePerms(); err != nil {
		return
	}
	if err = godotenv.Load(o.EnvFilePath); err != nil {
		err = fmt.Errorf(i18n.T("db_error_loading_env_file"), err)
	}
	return
}

func (o *Db) IsEnvFileExists() (ret bool) {
	_, err := os.Stat(o.EnvFilePath)
	ret = !os.IsNotExist(err)
	return
}

func (o *Db) SaveEnv(content string) (err error) {
	if err = os.MkdirAll(o.Dir, ConfigDirPerms); err != nil {
		return
	}
	if err = o.ensureSecureConfigDirPerms(); err != nil {
		return
	}
	err = os.WriteFile(o.EnvFilePath, []byte(content), EnvFilePerms)
	return
}

func (o *Db) LoadEnvMap() (map[string]string, error) {
	if err := o.ensureSecureEnvFilePerms(); err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	if !o.IsEnvFileExists() {
		return map[string]string{}, nil
	}

	values, err := godotenv.Read(o.EnvFilePath)
	if err != nil {
		return nil, fmt.Errorf(i18n.T("db_error_loading_env_file"), err)
	}

	return values, nil
}

func (o *Db) SaveEnvMap(values map[string]string) error {
	keys := make([]string, 0, len(values))
	for key, value := range values {
		if key == "" || value == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var content strings.Builder
	for _, key := range keys {
		content.WriteString(key)
		content.WriteString("=")
		content.WriteString(values[key])
		content.WriteString("\n")
	}

	return o.SaveEnv(content.String())
}

func (o *Db) UpdateEnvValues(updates map[string]*string) error {
	values, err := o.LoadEnvMap()
	if err != nil {
		return err
	}

	for key, value := range updates {
		if value == nil {
			continue
		}
		if *value == "" {
			delete(values, key)
			_ = os.Unsetenv(key)
			continue
		}

		values[key] = *value
		_ = os.Setenv(key, *value)
	}

	return o.SaveEnvMap(values)
}

func (o *Db) FilePath(fileName string) (ret string) {
	return filepath.Join(o.Dir, fileName)
}

func (o *Db) ensureSecureConfigDirPerms() error {
	if _, err := os.Stat(o.Dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return os.Chmod(o.Dir, ConfigDirPerms)
}

func (o *Db) ensureSecureEnvFilePerms() error {
	if _, err := os.Stat(o.EnvFilePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return os.Chmod(o.EnvFilePath, EnvFilePerms)
}

type DirectoryChange struct {
	Dir       string
	Timestamp time.Time
}
