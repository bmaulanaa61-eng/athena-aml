package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/investigation"
	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/errors"
)

type PostgresInvestigationRepository struct {
	DB *pgxpool.Pool
}

func NewPostgresInvestigationRepository(db *pgxpool.Pool) *PostgresInvestigationRepository {
	return &PostgresInvestigationRepository{DB: db}
}

func (r *PostgresInvestigationRepository) AddNote(ctx context.Context, tenantID, alertID string, req *investigation.AddNoteRequest) (*investigation.InvestigationNote, error) {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM alerts WHERE tenant_id = $1::uuid AND id = $2::uuid);`
	_ = r.DB.QueryRow(ctx, checkQuery, tenantID, alertID).Scan(&exists)
	if !exists {
		return nil, errors.NotFound(errors.CodeAlertNotFound, "Alert tidak ditemukan.")
	}

	insertQuery := `
		INSERT INTO investigation_notes (
			tenant_id, alert_id, author_id, note
		) VALUES (
			$1::uuid, $2::uuid, 
			COALESCE((SELECT id FROM users WHERE tenant_id = $1::uuid LIMIT 1), '00000000-0000-0000-0000-000000000000'::uuid),
			$3
		)
		RETURNING id::text, tenant_id::text, alert_id::text, author_id::text, note, created_at;
	`

	var n investigation.InvestigationNote
	var authorID string
	err := r.DB.QueryRow(ctx, insertQuery, tenantID, alertID, req.Note).Scan(
		&n.ID, &n.TenantID, &n.AlertID, &authorID, &n.Note, &n.CreatedAt,
	)
	if err != nil {
		return nil, errors.Internal(err, "Gagal menyimpan catatan investigasi ke database.")
	}
	n.UserID = &authorID
	n.Author = req.Author

	return &n, nil
}

func (r *PostgresInvestigationRepository) GetNotes(ctx context.Context, tenantID, alertID string) ([]investigation.InvestigationNote, error) {
	query := `
		SELECT 
			n.id::text, n.tenant_id::text, n.alert_id::text, n.author_id::text, COALESCE(u.email, 'Compliance Officer'), n.note, n.created_at
		FROM investigation_notes n
		LEFT JOIN users u ON n.author_id = u.id
		WHERE n.tenant_id = $1::uuid AND n.alert_id = $2::uuid
		ORDER BY n.created_at ASC;
	`

	list := []investigation.InvestigationNote{}

	rows, err := r.DB.Query(ctx, query, tenantID, alertID)
	if err != nil {
		return nil, errors.Internal(err, "Gagal mengambil riwayat catatan investigasi.")
	}
	defer rows.Close()

	for rows.Next() {
		var n investigation.InvestigationNote
		var authorID, authorEmail string
		if err := rows.Scan(&n.ID, &n.TenantID, &n.AlertID, &authorID, &authorEmail, &n.Note, &n.CreatedAt); err != nil {
			return nil, errors.Internal(err, "Gagal memproses baris catatan.")
		}
		n.UserID = &authorID
		n.Author = authorEmail
		list = append(list, n)
	}

	return list, nil
}

func (r *PostgresInvestigationRepository) ResolveCase(ctx context.Context, tenantID, alertID string, req *investigation.ResolveAlertRequest) (*investigation.InvestigationResolution, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, errors.Internal(err, "Gagal memulai transaksi resolusi.")
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	checkQuery := `SELECT status FROM alerts WHERE tenant_id = $1::uuid AND id = $2::uuid FOR UPDATE;`
	err = tx.QueryRow(ctx, checkQuery, tenantID, alertID).Scan(&currentStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound(errors.CodeAlertNotFound, "Alert tidak ditemukan.")
		}
		return nil, errors.Internal(err, "Gagal memeriksa status alert.")
	}

	insertResQuery := `
		INSERT INTO investigation_resolutions (
			tenant_id, alert_id, resolver_id, resolution, reason
		) VALUES (
			$1::uuid, $2::uuid,
			COALESCE((SELECT id FROM users WHERE tenant_id = $1::uuid LIMIT 1), '00000000-0000-0000-0000-000000000000'::uuid),
			$3, $4
		)
		RETURNING id::text, tenant_id::text, alert_id::text, resolver_id::text, resolution, reason, created_at;
	`

	var res investigation.InvestigationResolution
	var resolverID string
	err = tx.QueryRow(ctx, insertResQuery, tenantID, alertID, req.ResolutionType, req.Rationale).Scan(
		&res.ID, &res.TenantID, &res.AlertID, &resolverID, &res.ResolutionType, &res.Rationale, &res.CreatedAt,
	)
	if err != nil {
		return nil, errors.Internal(err, "Gagal menyimpan keputusan resolusi.")
	}
	res.UserID = &resolverID
	res.Resolver = req.Resolver

	updateAlertQuery := `UPDATE alerts SET status = 'RESOLVED', updated_at = NOW() WHERE id = $1::uuid;`
	_, err = tx.Exec(ctx, updateAlertQuery, alertID)
	if err != nil {
		return nil, errors.Internal(err, "Gagal memperbarui status alert menjadi RESOLVED.")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errors.Internal(err, "Gagal melakukan commit resolusi kasus.")
	}

	return &res, nil
}

func (r *PostgresInvestigationRepository) GetTimeline(ctx context.Context, tenantID, alertID string) (*investigation.CaseTimelineResponse, error) {
	var alertStatus string
	var alertCreatedAt time.Time
	err := r.DB.QueryRow(
		ctx,
		`SELECT status, created_at FROM alerts WHERE tenant_id = $1::uuid AND id = $2::uuid;`,
		tenantID, alertID,
	).Scan(&alertStatus, &alertCreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound(errors.CodeAlertNotFound, "Alert tidak ditemukan.")
		}
		return nil, errors.Internal(err, "Gagal mengambil data alert.")
	}

	timeline := []investigation.CaseTimelineItem{
		{
			EventType: "ALERT_TRIGGERED",
			Actor:     "ATHENA AML Engine",
			Detail:    "Peringatan anomali transaksi terdeteksi dan antrean investigasi dibuka.",
			Timestamp: alertCreatedAt,
		},
	}

	notes, _ := r.GetNotes(ctx, tenantID, alertID)
	for _, n := range notes {
		timeline = append(timeline, investigation.CaseTimelineItem{
			EventType: "NOTE_ADDED",
			Actor:     n.Author,
			Detail:    n.Note,
			Timestamp: n.CreatedAt,
		})
	}

	var resType, resRationale, resolverEmail string
	var resCreatedAt time.Time
	resErr := r.DB.QueryRow(
		ctx,
		`
		SELECT r.resolution, r.reason, COALESCE(u.email, 'Compliance Lead'), r.created_at 
		FROM investigation_resolutions r 
		LEFT JOIN users u ON r.resolver_id = u.id
		WHERE r.tenant_id = $1::uuid AND r.alert_id = $2::uuid 
		LIMIT 1;
		`,
		tenantID, alertID,
	).Scan(&resType, &resRationale, &resolverEmail, &resCreatedAt)

	if resErr == nil {
		timeline = append(timeline, investigation.CaseTimelineItem{
			EventType: "CASE_RESOLVED",
			Actor:     resolverEmail,
			Detail:    "Kasus ditutup sebagai " + resType + " dengan alasan: " + resRationale,
			Timestamp: resCreatedAt,
		})
	}

	return &investigation.CaseTimelineResponse{
		AlertID:  alertID,
		Status:   alertStatus,
		Timeline: timeline,
	}, nil
}
