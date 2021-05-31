SHELL=/bin/bash -o pipefail

generate_dashboards: clean
	for name in deck ghproxy hook plank prow sinker tide; do echo "Generating $${name} dashboard JSON"; jsonnet -J vendor/ -J lib/ "jsonnet/$${name}.jsonnet" > "dashboards_out/$${name}.json"; done

apply_configmaps:
	for name in deck ghproxy hook plank prow sinker tide; do kubectl -n prow-monitoring create configmap "grafana-dashboard-$${name}" --from-file="dashboards_out/$${name}.json"; done
	kubectl -n prow-monitoring create configmap "grafana-dashboard-test-coverage" --from-file="quality/test_coverage.json"
	kubectl -n prow-monitoring create configmap "grafana-dashboard-quality" --from-file="quality/quality.json"

clean:
	rm -f ./dashboards_out/*
