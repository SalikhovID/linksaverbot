package files

import (
	"encoding/gob"
	"errors"
	"fmt"
	"go-tg/lib/e"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"go-tg/storage"
)

const defaultPerm = 0774

type Storge struct {
	basePath string
}

func New(basePath string) Storge {
	return Storge{basePath: basePath}
}

func (s Storge) Save(page *storage.Page) (err error) {
	defer func ()  { err = e.WrapIfErr("cam't save page", err)} ()

	fPath := filepath.Join(s.basePath, page.UserName)

	if err:=os.MkdirAll(fPath, defaultPerm);err!=nil{
		return err
	}

	fileName, err := fileName(page)
	if err != nil {
		return err
	}

	fPath = filepath.Join(fPath, fileName)

	file, err := os.Create(fPath)

	if err != nil {
		return err
	}

	defer func() { _=file.Close()}()

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	return nil
}

func (s Storge) PickRandom(userName string) (page *storage.Page, err error) {
	defer func ()  { err = e.WrapIfErr("can't pick random page", err)} ()

	path := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(files))

	file := files[n]

	return s.decodePage(filepath.Join(path, file.Name()))
}

func (s Storge) Remove(p *storage.Page) error {
	fileName, err := fileName(p)
	if err != nil {
		return e.Wrap("can't remove page", err)
	}

	path := filepath.Join(s.basePath,p.UserName,fileName)

	if err := os.Remove(path); err != nil {
		return e.Wrap(fmt.Sprintf("can't remove page %s", path), err)
	}

	return nil
}

func (s Storge) IsExists(p *storage.Page) (bool, error) {
	fileName, err := fileName(p)
	if err != nil {
		return false, e.Wrap("can't check if file exists page", err)
	}

	path := filepath.Join(s.basePath,p.UserName,fileName)

	switch _, err = os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can't remove page %s", path)
		return false, e.Wrap(msg, err)
	}

	return true, nil
}

func (s Storge) decodePage(filePath string) (*storage.Page, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}
	defer func () {_=f.Close()}()
	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &p, nil
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}