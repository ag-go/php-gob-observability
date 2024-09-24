<?php

while (true) {
  MetricObserver::log("database_calls", "file", "file.php@L123", 0.142857);
  usleep(100 * 1000);
}

// echo "firewalled \n";
// MetricObserver::$address = '192.168.0.177';
// MetricObserver::$port = 7777;
// $time = microtime(true);
// for ($i = 0; $i < 10000000; $i++) {
//   MetricObserver::log("database_calls", "file", "file.php@L123", 0.142857);
// }
// echo (microtime(true) - $time) . "\n";

// echo "closed \n";
// MetricObserver::$address = 'localhost';
// MetricObserver::$port = 7778;
// $time = microtime(true);
// for ($i = 0; $i < 10000000; $i++) {
//   MetricObserver::log("database_calls", "file", "file.php@L123", 0.142857);
// }
// echo (microtime(true) - $time) . "\n";

// echo "open \n";
// MetricObserver::$address = 'localhost';
// MetricObserver::$port = 7777;
// $time = microtime(true);
// for ($i = 0; $i < 10000000; $i++) {
//   MetricObserver::log("database_calls", "file", "file.php@L123", 0.142857);
// }
// echo (microtime(true) - $time) . "\n";

class MetricObserver
{
  public static string $address = 'localhost';
  public static int $port = 7777;

  private static ?Socket $socket = null;
  private static bool $connected = false;
  private static int $connectAt = 0;

  public static function log(string $metricName, string $tagName, string $tagValue, float $duration)
  {
    if (!self::$socket) {
      self::$socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP) ?: null;
      self::$connected = false;
    }
    if (!self::$connected) {
      $now = time();
      if (self::$connectAt != $now) {
        self::$connectAt = $now;
        socket_set_option(self::$socket, SOL_SOCKET, SO_SNDTIMEO, ['sec' => 0, 'usec' => 1]);
        self::$connected = @socket_connect(self::$socket, self::$address, self::$port);
        socket_set_option(self::$socket, SOL_SOCKET, SO_SNDTIMEO, ['sec' => 30, 'usec' => 0]);
      }
    }
    if (self::$connected) {
      $line = json_encode([$metricName, $tagName, $tagValue, (string)$duration]);
      if (!@socket_write(self::$socket, $line . "\n", strlen($line) + 1)) {
        self::$socket = null;
        self::$connected = false;
      }
    }
  }
}
