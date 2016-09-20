package datastore

import (
	"fmt"

	"github.com/drone/drone/model"
	"github.com/russross/meddler"
)

func (db *datastore) GetRepo(id int64) (*model.Repo, error) {
	var repo = new(model.Repo)
	var err = meddler.Load(db, repoTable, repo, id)
	return repo, err
}

func (db *datastore) GetRepoName(name string) (*model.Repo, error) {
	var repo = new(model.Repo)
	var err = meddler.QueryRow(db, repo, rebind(repoNameQuery), name)
	return repo, err
}

func (db *datastore) GetRepoListOf(listof []*model.RepoLite) ([]*model.Repo, error) {
	var (
		err error
		repos = []*model.Repo{}
		toListRepoLite func([]*model.RepoLite) (string, []interface{})
		total = len(listof)
	)

	if total == 0 {
		return repos, nil
	}

	switch meddler.Default {
	case meddler.PostgreSQL:
		toListRepoLite = toListPosgres
	default:
		toListRepoLite = toList
	}

	pages := calculatePagination(total, maxRepoPage)

	for i := 0; i < pages; i++ {
		stmt, args := toListRepoLite(resizeList(listof, i, maxRepoPage))

		var tmpRepos []*model.Repo
		err = meddler.QueryAll(db, &tmpRepos, fmt.Sprintf(repoListOfQuery, stmt), args...)
		if err != nil {
			return nil, err
		}

		repos = append(repos, tmpRepos...)
	}
	return repos, nil
}

func (db *datastore) GetRepoCount() (int, error) {
	var count int
	var err = db.QueryRow(rebind(repoCountQuery)).Scan(&count)
	return count, err
}

func (db *datastore) CreateRepo(repo *model.Repo) error {
	return meddler.Insert(db, repoTable, repo)
}

func (db *datastore) UpdateRepo(repo *model.Repo) error {
	return meddler.Update(db, repoTable, repo)
}

func (db *datastore) DeleteRepo(repo *model.Repo) error {
	var _, err = db.Exec(rebind(repoDeleteStmt), repo.ID)
	return err
}

const repoTable = "repos"

const repoNameQuery = `
SELECT *
FROM repos
WHERE repo_full_name = ?
LIMIT 1;
`

const repoListOfQuery = `
SELECT *
FROM repos
WHERE repo_full_name IN (%s)
ORDER BY repo_name
`

const repoCountQuery = `
SELECT COUNT(*) FROM repos
`

const repoDeleteStmt = `
DELETE FROM repos
WHERE repo_id = ?
`
