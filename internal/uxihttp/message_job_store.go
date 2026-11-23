package uxihttp

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"errors"
	"os"
	"path/filepath"
	"regexp"
)

var validJobID = regexp.MustCompile(`^[a-f0-9]{32}$`)

type messageJobStore struct {
	dir string
}

func newMessageJobStore(root string) (*messageJobStore, error) {
	if root == "" {
		return nil, errors.New("empty job store root")
	}
	dir := filepath.Join(root, "message-jobs")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("job directory must be private and not a symlink")
	}
	return &messageJobStore{dir: dir}, nil
}

func (s *messageJobStore) save(job messageJob) error {
	if !validJobID.MatchString(job.ID) {
		return errors.New("invalid job ID")
	}
	if job.SessionID == "" || job.CreatedAt.IsZero() || !validJobStatus(job.Status) {
		return errors.New("invalid job record")
	}
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	if len(data) > 1<<20 { return errors.New("job record exceeds size limit") }
	path := filepath.Join(s.dir, job.ID+".json")
	tmp, err := os.CreateTemp(s.dir, ".job-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil { return err }
	d, err := os.Open(s.dir)
	if err != nil { return err }
	defer d.Close()
	return d.Sync()
}

func (s *messageJobStore) load(id string) (messageJob, error) {
	if !validJobID.MatchString(id) {
		return messageJob{}, errors.New("invalid job ID")
	}

	path := filepath.Join(s.dir, id+".json")
	info, err := os.Lstat(path)
	if err != nil { return messageJob{}, err }
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || info.Size() > 1<<20 { return messageJob{}, errors.New("unsafe job record") }
	file, err := os.Open(path)
	if err != nil { return messageJob{}, err }
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil { return messageJob{}, err }
	if len(data) > 1<<20 {
		return messageJob{}, errors.New("job record exceeds size limit")
	}

	var job messageJob
	if err := json.Unmarshal(data, &job); err != nil {
		return messageJob{}, err
	}
	if job.ID != id || job.SessionID == "" || job.CreatedAt.IsZero() || !validJobStatus(job.Status) {
		return messageJob{}, errors.New("job record ID mismatch")
	}

	return job, nil
}

func validJobStatus(status string) bool {
	switch status {
	case "running", "completed", "failed", "interrupted":
		return true
	}
	return false
}

func (s *messageJobStore) list() ([]messageJob, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil { return nil, err }
	if len(entries) > 10000 { return nil, errors.New("too many persisted jobs") }
	jobs := make([]messageJob, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".job-") && strings.HasSuffix(name, ".tmp") { continue }
		if !strings.HasSuffix(name, ".json") { return nil, fmt.Errorf("unexpected job file") }
		id := strings.TrimSuffix(name, ".json")
		job, err := s.load(id)
		if err != nil { return nil, fmt.Errorf("invalid persisted job: %w", err) }
		jobs = append(jobs, job)
	}
	return jobs, nil
}
