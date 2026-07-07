package resourceimport

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
)

type importFixture struct {
	svc          *Service
	resourceRepo *fake.ResourceFake
	scheduler    *fake.SchedulerFake
	channelRepo  *fake.NotificationChannelFake
	tagsRepo     *fake.TagsFake
}

func newImportFixture(t *testing.T) *importFixture {
	t.Helper()
	resourceRepo := fake.NewResourceFake()
	incidentRepo := fake.NewIncidentFake()
	tagsRepo := fake.NewTagsFake()
	scheduler := fake.NewSchedulerFake()
	monitoring := fake.NewMonitoringActivityFake()
	channelRepo := fake.NewNotificationChannelFake()
	componentRepo := fake.NewComponentFake()
	enrichment := service.NewEnrichmentService(30 * time.Second)
	componentSvc := service.NewComponentService(componentRepo, resourceRepo, channelRepo)
	resourceSvc := service.NewResourceService(resourceRepo, incidentRepo, tagsRepo, channelRepo, scheduler, monitoring, enrichment, componentSvc)

	return &importFixture{
		svc:          NewService(resourceSvc, componentRepo, channelRepo),
		resourceRepo: resourceRepo,
		scheduler:    scheduler,
		channelRepo:  channelRepo,
		tagsRepo:     tagsRepo,
	}
}

// seedTag pre-creates a tag with an ID. The in-memory TagsFake requires a
// non-empty ID (the real sqlc Create wrapper assigns one via EnsureID).
func (f *importFixture) seedTag(t *testing.T, name string) {
	t.Helper()
	if err := f.tagsRepo.Create(context.Background(), &domain.Tags{
		Base: domain.Base{ID: "tag-" + name},
		Name: name,
	}); err != nil {
		t.Fatalf("seed tag: %v", err)
	}
}

func (f *importFixture) seedChannel(t *testing.T, name string) {
	t.Helper()
	if err := f.channelRepo.Create(context.Background(), &domain.NotificationChannel{
		Base: domain.Base{ID: "chan-" + name},
		Name: name,
		Type: domain.NotificationChannelTypeSMTP,
	}); err != nil {
		t.Fatalf("seed channel: %v", err)
	}
}

const mixedManifest = `
version: 1
defaults:
  interval: 60
  timeout: 10
resources:
  - name: Site
    type: http
    target: https://example.com
    tags: [prod, web]
    component: Website
    notification_channels: [ops-email]
  - name: DB
    type: tcp
    target: example.com:443
  - name: Beat
    type: heartbeat
    heartbeat_interval: 3600
    heartbeat_grace: 300
`

func TestImport_CreatesAndSchedules(t *testing.T) {
	f := newImportFixture(t)
	f.seedChannel(t, "ops-email")
	f.seedTag(t, "prod")
	f.seedTag(t, "web")

	report, err := f.svc.Import(context.Background(), []byte(mixedManifest), dtoV1.ImportOptions{})
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if report.Created != 3 {
		t.Fatalf("created = %d, want 3 (rows: %+v)", report.Created, report.Rows)
	}

	all, _ := f.resourceRepo.List(context.Background(), 100, 0)
	if len(all) != 3 {
		t.Fatalf("resources persisted = %d, want 3", len(all))
	}
	for _, r := range all {
		if !f.scheduler.IsScheduled(r.ID) {
			t.Fatalf("resource %q was not scheduled", r.Name)
		}
	}
}

func TestImport_DryRunWritesNothing(t *testing.T) {
	f := newImportFixture(t)
	f.seedChannel(t, "ops-email")

	report, err := f.svc.Import(context.Background(), []byte(mixedManifest), dtoV1.ImportOptions{DryRun: true})
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	if !report.DryRun || report.Created != 0 {
		t.Fatalf("dry-run report wrong: %+v", report)
	}
	all, _ := f.resourceRepo.List(context.Background(), 100, 0)
	if len(all) != 0 {
		t.Fatalf("dry-run wrote %d resources, want 0", len(all))
	}
}

func TestImport_InvalidRowBlocksAll(t *testing.T) {
	f := newImportFixture(t)
	// Second row is invalid (keyword missing keyword).
	manifest := `
version: 1
defaults: { interval: 60, timeout: 10 }
resources:
  - name: Good
    type: http
    target: https://example.com
  - name: Bad
    type: keyword
    target: https://example.com
`
	report, err := f.svc.Import(context.Background(), []byte(manifest), dtoV1.ImportOptions{})
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("err = %v, want ErrValidationFailed", err)
	}
	if report == nil || report.Failed != 1 {
		t.Fatalf("report failed count wrong: %+v", report)
	}
	all, _ := f.resourceRepo.List(context.Background(), 100, 0)
	if len(all) != 0 {
		t.Fatalf("all-or-nothing violated: %d resources created", len(all))
	}
}

func TestImport_DuplicatePolicy(t *testing.T) {
	ctx := context.Background()

	// skip: existing "Dup" skipped, "New" created.
	f := newImportFixture(t)
	_, _ = f.resourceRepo.Create(ctx, &domain.Resource{Base: domain.Base{ID: "existing"}, Name: "Dup", Type: domain.ResourceHTTP})
	manifest := `
version: 1
defaults: { interval: 60, timeout: 10 }
resources:
  - name: Dup
    type: http
    target: https://example.com
  - name: New
    type: http
    target: https://example.org
`
	report, err := f.svc.Import(ctx, []byte(manifest), dtoV1.ImportOptions{DuplicatePolicy: dtoV1.DuplicatePolicySkip})
	if err != nil {
		t.Fatalf("skip import failed: %v", err)
	}
	if report.Created != 1 || report.Skipped != 1 {
		t.Fatalf("skip counts wrong: %+v", report)
	}

	// error: existing "Dup" fails the whole import, nothing new created.
	f2 := newImportFixture(t)
	_, _ = f2.resourceRepo.Create(ctx, &domain.Resource{Base: domain.Base{ID: "existing"}, Name: "Dup", Type: domain.ResourceHTTP})
	_, err = f2.svc.Import(ctx, []byte(manifest), dtoV1.ImportOptions{DuplicatePolicy: dtoV1.DuplicatePolicyError})
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("error policy err = %v, want ErrValidationFailed", err)
	}
	all, _ := f2.resourceRepo.List(ctx, 100, 0)
	if len(all) != 1 { // only the pre-seeded existing one
		t.Fatalf("error policy created extra resources: %d", len(all))
	}
}

// stubGateway fails CreateResource on the Nth call to exercise compensation.
type stubGateway struct {
	failOn  int
	calls   int
	created []string
	deleted []string
}

func (s *stubGateway) CreateResource(_ context.Context, p *dto.CreateResourcePayload) (*domain.Resource, error) {
	s.calls++
	if s.calls == s.failOn {
		return nil, errors.New("boom")
	}
	id := "res-" + p.Name
	s.created = append(s.created, id)
	return &domain.Resource{Base: domain.Base{ID: id}, Name: p.Name}, nil
}

func (s *stubGateway) DeleteResource(_ context.Context, id string) error {
	s.deleted = append(s.deleted, id)
	return nil
}

func (s *stubGateway) ListAll(_ context.Context) ([]*domain.Resource, error) { return nil, nil }

func TestImport_CompensatesOnFailure(t *testing.T) {
	stub := &stubGateway{failOn: 2}
	svc := NewService(stub, fake.NewComponentFake(), fake.NewNotificationChannelFake())

	manifest := `
version: 1
defaults: { interval: 60, timeout: 10 }
resources:
  - name: A
    type: http
    target: https://a.example.com
  - name: B
    type: http
    target: https://b.example.com
  - name: C
    type: http
    target: https://c.example.com
`
	_, err := svc.Import(context.Background(), []byte(manifest), dtoV1.ImportOptions{})
	if err == nil {
		t.Fatal("expected import failure")
	}
	// Row A created then rolled back; row B failed; C never attempted.
	if len(stub.created) != 1 || len(stub.deleted) != 1 || stub.deleted[0] != stub.created[0] {
		t.Fatalf("compensation wrong: created=%v deleted=%v", stub.created, stub.deleted)
	}
}
