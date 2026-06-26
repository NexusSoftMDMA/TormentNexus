#include <Servo.h>
#include "config.h"
#define LED_PIN 13
#define BAUD_RATE 9600

// Motor controller class
class MotorController {
public:
    explicit MotorController(int pin) : pin_(pin) {}

    void start() {
        analogWrite(pin_, 255);
    }

    void stop() {
        analogWrite(pin_, 0);
    }

private:
    int pin_;
};

struct SensorReading {
    float temperature;
    float humidity;
};

enum State {
    STATE_IDLE,
    STATE_RUNNING,
    STATE_ERROR,
};

void setup() {
    Serial.begin(BAUD_RATE);
    pinMode(LED_PIN, OUTPUT);
}

void loop() {
    digitalWrite(LED_PIN, HIGH);
    delay(1000);
    digitalWrite(LED_PIN, LOW);
    delay(1000);
}

float readTemperature(int sensorPin) {
    int raw = analogRead(sensorPin);
    return raw * 0.48828125;
}
