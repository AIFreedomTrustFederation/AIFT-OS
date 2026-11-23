package uxihttp

import (
	"encoding/json"
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
	return &messageJobStore{dir: dir}, nil
}

func (s *messageJobStore) save(job messageJob) error {
	if !validJobID.MatchString(job.ID) {
		return errors.New("invalid job ID")
	}
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
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
	return os.Rename(tmp.Name(), path)
}

func (s *messageJobStore) load(id string) (messageJob, error) {
	if !validJobID.MatchString(id) {
		return messageJob{}, errors.New("invalid job ID")
	}

	path := filepath.Join(s.dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return messageJob{}, err
	}
	if len(data) > 1<<20 {
		return messageJob{}, errors.New("job record exceeds size limit")
	}

	var job messageJob
	if err := json.Unmarshal(data, &job); err != nil {
		return messageJob{}, err
	}
	if job.ID != id {
		return messageJob{}, errors.New("job record ID mismatch")
	}

	return job, nil
}
