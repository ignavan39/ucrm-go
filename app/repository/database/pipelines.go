package database

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/ignavan39/ucrm-go/app/models"
)

func (r *DbService) AddPipeline(name string, dashboardId string, order int) (*models.Pipeline, error) {
	pipeline := &models.Pipeline{}

	row := sq.Insert("pipelines").Columns("name", "dashboard_id", `"order"`).
		Values(name, dashboardId, order).
		Suffix(`returning id,name,"order",dashboard_id,updated_at`).
		RunWith(r.pool.Write()).PlaceholderFormat(sq.Dollar).QueryRow()
	if err := row.Scan(&pipeline.Id, &pipeline.Name, &pipeline.Order, &pipeline.DashboardId, &pipeline.UpdatedAt); err != nil {
		return nil, err
	}

	return pipeline, nil
}

func (r *DbService) GetOnePipeline(pipelineId string) (*models.Pipeline, error) {
	pipeline := &models.Pipeline{}

	row := sq.Select("id", "name", `"order"`, "dashboard_id", "updated_at").
		From("pipelines").
		Where(sq.Eq{"id": pipelineId}).
		RunWith(r.pool.Write()).
		PlaceholderFormat(sq.Dollar).
		QueryRow()
	if err := row.Scan(&pipeline.Id, &pipeline.Name, &pipeline.Order, &pipeline.DashboardId, &pipeline.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return pipeline, nil
}

func (r *DbService) GetAccessPipelineById(pipelineId string, userId string, accessType string) (bool, error) {
	var id string

	builder := sq.Select("p.id").
		From("pipelines p").
		Where(sq.Eq{"p.id": pipelineId}).
		InnerJoin("dashboards d on p.dashboard_id = d.id").
		InnerJoin("dashboards_user du on d.id = du.dashboard_id").
		Where(sq.Eq{"p.id": pipelineId, "du.user_id": userId})
	if accessType == "r" {
		builder.Where(sq.Or{sq.Eq{"du.access": accessType}, sq.Eq{"du.access": "rw"}})
	} else {
		builder.Where(sq.Eq{"du.access": accessType})
	}

	row := builder.RunWith(r.pool.Read()).
		PlaceholderFormat(sq.Dollar).
		QueryRow()
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *DbService) GetAllPipelines(dashboardId string) ([]models.Pipeline, error) {
	pipelines := []models.Pipeline{}

	rows, err := sq.Select("id", "name", `"order"`, "dashboard_id", "updated_at").
		From("pipelines").
		Where(sq.Eq{"dashboard_id": dashboardId}).
		OrderBy(`"order"`).
		RunWith(r.pool.Read()).
		PlaceholderFormat(sq.Dollar).
		Query()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var p models.Pipeline
		if err := rows.Scan(&p.Id, &p.Name, &p.Order, &p.DashboardId, &p.UpdatedAt); err != nil {
			return nil, err
		}
		pipelines = append(pipelines, p)
	}

	return pipelines, nil
}

func (r *DbService) GetAllPipelinesByPipeline(pipelineId string) ([]models.Pipeline, error) {
	pipelines := []models.Pipeline{}

	var dashboardId string
	row := sq.Select("dashboard_id").
		From("pipelines").
		Where(sq.Eq{"id": pipelineId}).
		RunWith(r.pool.Read()).
		PlaceholderFormat(sq.Dollar).
		QueryRow()
	if err := row.Scan(&dashboardId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]models.Pipeline, 0), nil
		}
		return make([]models.Pipeline, 0), err
	}

	rows, err := sq.Select("id", "name", `"order"`, "dashboard_id", "updated_at").
		From("pipelines").
		Where(sq.Eq{"dashboard_id": dashboardId}).
		OrderBy(`"order"`).
		RunWith(r.pool.Read()).
		PlaceholderFormat(sq.Dollar).
		Query()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var p models.Pipeline
		if err := rows.Scan(&p.Id, &p.Name, &p.Order, &p.DashboardId, &p.UpdatedAt); err != nil {
			return nil, err
		}
		pipelines = append(pipelines, p)
	}

	return pipelines, nil
}

func (r *DbService) UpdatePipelineName(pipelineId string, name string) error {
	_, err := sq.Update("pipelines").
		Set("name", name).
		Where(sq.Eq{"id": pipelineId}).
		RunWith(r.pool.Write()).
		PlaceholderFormat(sq.Dollar).
		Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return err
}

func (r *DbService) DeletePipelineById(pipelineId string) error {
	_, err := sq.Delete("pipelines cascade").
		Where(sq.Eq{"id": pipelineId}).
		RunWith(r.pool.Write()).
		PlaceholderFormat(sq.Dollar).
		Exec()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return err
}

func (r *DbService) UpdateOrderForPipeline(pipelineId string, order int) error {

	_, err :=
		sq.Update("pipelines p").
			Set(`"order"`, order).
			Where(sq.Eq{"id": pipelineId}).
			RunWith(r.pool.Write()).
			PlaceholderFormat(sq.Dollar).
			Exec()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return nil
}
