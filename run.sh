#!/bin/bash

set -eou pipefail

TARGET_ENV="${TARGET_ENV:-development}"

trap 'tear_down;kill $(jobs -p); exit 0' EXIT

tear_down() {
    if [[ "$TARGET_ENV" = "development" ]]; then
        echo -e "\nUndeploying dev manifests"
    fi
}

scrape_cluster_logging_es_metrics() {
    source .bingo/variables.env
    (
        $PROMETHEUS --log.level=warn --config.file=./config/prometheus/config.yaml --storage.tsdb.path="$(mktemp -d)";
    ) &
}

generate_report() {
    source .bingo/variables.env

    for f in $REPORT_DIR/*.gnuplot; do
        gnuplot -e "set term png; set output '$f.png'" "$f"
    done

    cp ./reports/README.template $REPORT_DIR/README.md
    sed -i "s/{{TARGET_ENV}}/$TARGET_ENV/i" $REPORT_DIR/README.md
    $EMBEDMD -w $REPORT_DIR/README.md
}


bench() {
    if [[ "$TARGET_ENV" = "development" ]]; then
        echo "Deploying dev manifests"
    fi

    echo -e "\nScrape metrics from Loki deployments"
    scrape_cluster_logging_es_metrics

    source .bingo/variables.env

    echo -e "\nRun benchmarks"
    $GINKGO -v ./benchmarks

    echo -e "\nGenerate benchmark report"
    generate_report
}

bench

exit $?
