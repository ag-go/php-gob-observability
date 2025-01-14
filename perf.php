<?php

include 'MetricObserver.php';

echo "MetricObserver::log() to a closed port:\n";
MetricObserver::$port = 7778;
$time = microtime(true);
for ($i = 0; $i < 1000000; $i++) {
    MetricObserver::log("database_calls", "file", "file.php@L123", 0.142857);
}
echo sprintf("%.2f usec\n", microtime(true) - $time);

sleep(2);

echo "MetricObserver::log() to an open port:\n";
MetricObserver::$port = 7777;
$time = microtime(true);
for ($i = 0; $i < 1000000; $i++) {
    MetricObserver::log("database_calls", "file", "file.php@L123", 0.142857);
}
echo sprintf("%.2f usec\n", microtime(true) - $time);

sleep(2);

echo "sprintf('%.2f', 3.14) for comparison:\n";
MetricObserver::$port = 7777;
$time = microtime(true);
for ($i = 0; $i < 1000000; $i++) {
    sprintf("%.2f", 3.14);
}
echo sprintf("%.2f usec\n", microtime(true) - $time);
