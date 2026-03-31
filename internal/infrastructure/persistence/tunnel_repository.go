package persistence

import (
	tunneldomain "podux/internal/domain/tunnel"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type TunnelRepository struct {
	app core.App
}

func NewTunnelRepository(app core.App) *TunnelRepository {
	return &TunnelRepository{app: app}
}

func (r *TunnelRepository) UpdateStatusByServerID(serverID string, status tunneldomain.TunnelStatus) error {
	_, err := r.app.DB().
		Update("th_tunnels", dbx.Params{"status": string(status)}, dbx.HashExp{"serverId": serverID}).
		Execute()
	if err != nil {
		r.app.Logger().Error("Failed to update tunnel status by server id", "serverID", serverID, "error", err)
	}
	return err
}

func (r *TunnelRepository) UpdateStatusByIntegrationID(integrationID string, status tunneldomain.TunnelStatus) error {
	_, err := r.app.DB().
		Update("th_tunnels", dbx.Params{"status": string(status)}, dbx.HashExp{"integrationId": integrationID}).
		Execute()
	if err != nil {
		r.app.Logger().Error("Failed to update tunnel status by integration id", "integrationID", integrationID, "error", err)
	}
	return err
}

func (r *TunnelRepository) ResetAllStatus() error {
	_, err := r.app.DB().
		Update("th_tunnels", dbx.Params{"status": string(tunneldomain.TunnelStatusInactive)},
			dbx.Not(dbx.HashExp{"status": string(tunneldomain.TunnelStatusInactive)})).
		Execute()
	if err != nil {
		r.app.Logger().Error("Failed to reset all tunnel status", "error", err)
	}
	return err
}

func (r *TunnelRepository) CountByStatus(status tunneldomain.TunnelStatus) (int64, error) {
	var count int64
	err := r.app.DB().
		Select("COUNT(*)").
		From("th_tunnels").
		Where(dbx.HashExp{"status": string(status)}).
		Row(&count)
	if err != nil {
		r.app.Logger().Error("Failed to count tunnels by status", "status", status, "error", err)
	}
	return count, err
}
