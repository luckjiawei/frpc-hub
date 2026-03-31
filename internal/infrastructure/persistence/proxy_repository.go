package persistence

import (
	"podux/internal/domain/proxy"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type ProxyRepository struct {
	app core.App
}

func NewProxyRepository(app core.App) *ProxyRepository {
	return &ProxyRepository{app: app}
}

func (r *ProxyRepository) CountByStatus(status proxy.ProxyStatus) (int64, error) {
	var count int64
	err := r.app.DB().
		Select("COUNT(*)").
		From("fh_proxies").
		Where(dbx.HashExp{"status": string(status)}).
		Row(&count)

	if err != nil {
		r.app.Logger().Error("Failed to get proxy count by status", "error", err)
		return 0, err
	}
	return count, nil
}

func (r *ProxyRepository) CountByBootStatus(status proxy.ProxyBootStatus) (int64, error) {
	var count int64
	err := r.app.DB().
		Select("COUNT(*)").
		From("fh_proxies").
		Where(dbx.HashExp{"bootStatus": string(status)}).
		Row(&count)

	if err != nil {
		r.app.Logger().Error("Failed to get proxy count by bootStatus", "error", err)
		return 0, err
	}
	return count, nil
}

func (r *ProxyRepository) CountByType() (map[string]int64, error) {
	type proxyTypeCount struct {
		ProxyType string `db:"proxyType"`
		Count     int64  `db:"count"`
	}

	var results []proxyTypeCount
	err := r.app.DB().
		Select("proxyType", "COUNT(*) as count").
		From("fh_proxies").
		GroupBy("proxyType").
		All(&results)

	if err != nil {
		r.app.Logger().Error("Failed to get proxy counts by type", "error", err)
		return nil, err
	}

	counts := make(map[string]int64)
	for _, result := range results {
		counts[result.ProxyType] = result.Count
	}
	return counts, nil
}

func (r *ProxyRepository) UpdateBootStatus(id string, status proxy.ProxyBootStatus) error {
	r.app.Logger().Info("Update proxy boot status", "id", id, "status", status)

	result, err := r.app.DB().
		Update("fh_proxies", dbx.Params{
			"bootStatus": string(status),
		}, dbx.HashExp{"id": id}).
		Execute()

	if err != nil {
		r.app.Logger().Error("Failed to update proxy boot status", "id", id, "error", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	r.app.Logger().Info("Proxy boot status updated successfully", "id", id, "status", status, "rowsAffected", rowsAffected)
	return nil
}

func (r *ProxyRepository) UpdateBootStatusByServerID(serverID string, status proxy.ProxyBootStatus) error {
	r.app.Logger().Info("Update proxy boot status by server id", "id", serverID, "status", status)

	// Route through th_tunnels: fh_proxies.tunnelId → th_tunnels.serverId
	type tunnelRow struct {
		ID string `db:"id"`
	}
	var tunnels []tunnelRow
	if err := r.app.DB().
		Select("id").
		From("th_tunnels").
		Where(dbx.HashExp{"serverId": serverID}).
		All(&tunnels); err != nil {
		r.app.Logger().Error("Failed to find tunnels for server", "serverID", serverID, "error", err)
		return err
	}
	if len(tunnels) == 0 {
		return nil
	}

	tunnelIDs := make([]any, len(tunnels))
	for i, t := range tunnels {
		tunnelIDs[i] = t.ID
	}

	result, err := r.app.DB().
		Update("fh_proxies", dbx.Params{
			"bootStatus": string(status),
		}, dbx.In("tunnelId", tunnelIDs...)).
		Execute()

	if err != nil {
		r.app.Logger().Error("Failed to update proxy boot status by server id", "serverID", serverID, "error", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	r.app.Logger().Info("Proxy boot status updated successfully", "serverID", serverID, "status", status, "count", rowsAffected)
	return nil
}

func (r *ProxyRepository) UpdateBootStatusByIntegrationID(integrationID string, status proxy.ProxyBootStatus) error {
	r.app.Logger().Info("Update proxy boot status by integration id", "id", integrationID, "status", status)

	// fh_proxies.integrationId has been replaced by tunnelId (→ th_tunnels).
	// Find tunnel IDs for this integration, then batch-update proxies.
	type tunnelRow struct {
		ID string `db:"id"`
	}
	var tunnels []tunnelRow
	if err := r.app.DB().
		Select("id").
		From("th_tunnels").
		Where(dbx.HashExp{"integrationId": integrationID}).
		All(&tunnels); err != nil {
		r.app.Logger().Error("Failed to find tunnels for integration", "integrationID", integrationID, "error", err)
		return err
	}
	if len(tunnels) == 0 {
		return nil
	}

	tunnelIDs := make([]any, len(tunnels))
	for i, t := range tunnels {
		tunnelIDs[i] = t.ID
	}

	result, err := r.app.DB().
		Update("fh_proxies", dbx.Params{
			"bootStatus": string(status),
		}, dbx.In("tunnelId", tunnelIDs...)).
		Execute()

	if err != nil {
		r.app.Logger().Error("Failed to update proxy boot status by integration id", "integrationID", integrationID, "error", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	r.app.Logger().Info("Proxy boot status updated successfully", "integrationID", integrationID, "status", status, "count", rowsAffected)
	return nil
}

func (r *ProxyRepository) ResetAllBootStatus() error {
	r.app.Logger().Info("Resetting all proxy boot status to offline")

	result, err := r.app.DB().
		Update("fh_proxies", dbx.Params{
			"bootStatus": string(proxy.ProxyBootStatusOffline),
		}, dbx.Not(dbx.HashExp{
			"bootStatus": string(proxy.ProxyBootStatusOffline),
		})).
		Execute()

	if err != nil {
		r.app.Logger().Error("Failed to reset proxy boot status", "error", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	r.app.Logger().Info("Proxy boot status reset completed", "count", rowsAffected)
	return nil
}

func (r *ProxyRepository) FindByServerID(serverID string) ([]proxy.Proxy, error) {
	var proxies []proxy.Proxy
	err := r.app.DB().
		Select("fp.*").
		From("fh_proxies fp").
		InnerJoin("th_tunnels tt", dbx.NewExp("tt.id = fp.tunnelId")).
		Where(dbx.HashExp{"tt.serverId": serverID}).
		All(&proxies)

	if err != nil {
		r.app.Logger().Error("Failed to find proxies by server id", "serverID", serverID, "error", err)
		return nil, err
	}
	return proxies, nil
}

func (r *ProxyRepository) FindEnabledByServerID(serverID string) ([]proxy.Proxy, error) {
	var proxies []proxy.Proxy
	err := r.app.DB().
		Select("fp.*").
		From("fh_proxies fp").
		InnerJoin("th_tunnels tt", dbx.NewExp("tt.id = fp.tunnelId")).
		Where(dbx.And(
			dbx.HashExp{"tt.serverId": serverID},
			dbx.HashExp{"fp.status": string(proxy.ProxyStatusEnabled)},
		)).
		All(&proxies)

	if err != nil {
		r.app.Logger().Error("Failed to find enabled proxies by server id", "serverID", serverID, "error", err)
		return nil, err
	}
	return proxies, nil
}
