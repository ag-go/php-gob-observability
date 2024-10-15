# php-observability

A code base to showcase high frequency logging in PHP and aggregating into metrics using Go. Read the blog post:

[https://tqdev.com/2024-high-frequency-metrics-in-php-using-tcp-sockets](https://tqdev.com/2024-high-frequency-metrics-in-php-using-tcp-sockets)

and

[https://tqdev.com/2024-distributed-metrics-in-php-using-go-and-gob](https://tqdev.com/2024-distributed-metrics-in-php-using-go-and-gob)

### Requirements

- PHP 7.4 and higher

### Usage

To run the server:

    go run .

In bash run:

    for run in {1..100}; do php writer.php & done

And to stop:

    killall php

Now observe the metrics:

http://localhost:8080/

NB: The metrics are Prometheus compatible and follow the [OpenMetrics specification](https://github.com/OpenObservability/OpenMetrics/).

The metrics are also served in the Gob binary format on:

http://localhost:9999/

This endpoint can be scraped by other tools such as [metrics-db-importer](https://github.com/mevdschee/metrics-db-importer) or this tool itself to aggregate the metrics of multiple nodes (using the "-scrape" parameter).

### Example metrics

Here is an example of published metrics:

    # HELP database_calls_seconds A summary of the database calls.
    # TYPE database_calls_seconds summary
    database_calls_seconds_count{file="src/Controller/UserController.php@L123"} 6630
    database_calls_seconds_sum{file="src/Controller/UserController.php@L123"} 947.142
    # HELP database_calls_total_seconds A histogram of the database calls.
    # TYPE database_calls_total_seconds histogram
    database_calls_total_seconds_bucket{le="0.005"} 0
    database_calls_total_seconds_bucket{le="0.01"} 0
    database_calls_total_seconds_bucket{le="0.025"} 0
    database_calls_total_seconds_bucket{le="0.05"} 0
    database_calls_total_seconds_bucket{le="0.1"} 0
    database_calls_total_seconds_bucket{le="0.25"} 6630
    database_calls_total_seconds_bucket{le="0.5"} 6630
    database_calls_total_seconds_bucket{le="1"} 6630
    database_calls_total_seconds_bucket{le="2.5"} 6630
    database_calls_total_seconds_bucket{le="5"} 6630
    database_calls_total_seconds_bucket{le="10"} 6630
    database_calls_total_seconds_bucket{le="+Inf"} 6630
    database_calls_total_seconds_sum 947.142
    database_calls_total_seconds_count 6630

Enjoy!