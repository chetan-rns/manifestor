package eventlistener

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bigkevmcd/manifestor/pkg/layout"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ciFilter = "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == '%s'"

func GenerateEventListener(n string, m *layout.Manifest) *triggersv1.EventListener {
	return &triggersv1.EventListener{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventListener",
			APIVersion: "triggers.tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
		Spec: triggersv1.EventListenerSpec{
			Triggers: makeEventListenerTriggers(m),
		},
	}
}

func makeEventListenerTriggers(m *layout.Manifest) []triggersv1.EventListenerTrigger {
	l, _ := parseManifest(m)
	triggers := make([]triggersv1.EventListenerTrigger, len(l))
	for i, s := range l {
		triggers[i] = createListenerTrigger(s.Name, ciFilter, s.RepoName, s.CI, s.CD)
	}
	return triggers
}

type serviceVisitor struct {
	services []service
}

// TODO: DO NOT PANIC!
func extractRepo(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	parts := strings.Split(parsed.Path, "/")
	return fmt.Sprintf("%s/%s", parts[1], parts[2])
}

func (ev *serviceVisitor) Service(env *layout.Environment, app *layout.Application, svc *layout.Service) error {
	if svc.SourceURL == "" {
		return nil
	}
	ev.services = append(ev.services, service{extractRepo(svc.SourceURL), fmt.Sprintf("%s-%s", app.Name, svc.Name), env.Name, env.Pipelines.Integration, env.Pipelines.Deployment})
	return nil
}

func (ev *serviceVisitor) Application(env *layout.Environment, app *layout.Application) error {
	return nil
}

func (ev *serviceVisitor) Environment(env *layout.Environment) error {
	return nil
}

type service struct {
	RepoName string
	Name     string
	Env      string
	CI       string
	CD       string
}

func parseManifest(m *layout.Manifest) ([]service, error) {
	l := &serviceVisitor{services: []service{}}
	m.Walk(l)
	return l.services, nil
}