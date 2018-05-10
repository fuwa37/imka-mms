#include <ESP8266WiFi.h>
#include "DHT.h"
#define DHTPIN 2
#define DHTTYPE DHT22
WiFiClient client;
DHT dht(DHTPIN,DHTTYPE);
const char* ssid = "bimasakti";    // Enter SSID here
const char* password = "paladin4";  //Enter Password here

const String server = "imka-mms.herokuapp.com";
String getH;
String getT;
String getStr;
#define buzzer 15 
#define led1 14
#define led2 12
#define led3 13
void setup() {
  // put your setup code here, to run once:
  Serial.begin(9600);
  delay(1000);
  Serial.println("Connecting to :");
  Serial.println(ssid);
  WiFi.begin(ssid,password);
  pinMode(led1, OUTPUT);pinMode(led2, OUTPUT);pinMode(led3, OUTPUT);pinMode(buzzer, OUTPUT);
  
  while (WiFi.status()!=WL_CONNECTED){
    delay(500);
    Serial.print(".");
  }
  Serial.println("");
  Serial.println("WiFi connected");
  Serial.println("DHTxx test!");
  dht.begin();
  digitalWrite(led1,LOW); digitalWrite(led2,LOW); digitalWrite(led3,LOW); digitalWrite(buzzer,LOW);
}

void loop() {
  float h=dht.readHumidity();
  float t=dht.readTemperature();
  if (isnan(h) || isnan(t)) {
    Serial.println("Failed to read from DHT sensor!");
    return;
  }
  float hic=dht.computeHeatIndex(t,h,false);
  Serial.print("Humidity: ");
  Serial.print(h);
  Serial.print(" %\t");
  Serial.print("Temperature: ");
  Serial.print(t);
  Serial.print("Heat index: ");
  Serial.print(hic);
  Serial.println(" *C ");
  getH="";
  getT="";
  getH=getH+String((int)h);
  getT=getT+String((int)t);
  getStr="/input?suhu="+getT+"&klb="+getH;
  Serial.println("Accessing: ");
  Serial.println(server+getStr);
  delay(2000);
  
  if (client.connect(server, 80)) {
    
    Serial.println("connected");
    client.println("GET "+getStr+" HTTP/1.1");
    client.println("Host: "+server);
    client.println("Connection: close");
    client.println();
    String testStr = "";  // This version takes 3512 bytes    
    int len;
    String cekStr;
    while(client.connected()){
      char c = client.read();
      Serial.print(c);
      testStr=testStr+c;
    }
    len=testStr.length();
    Serial.println(len);
    cekStr=testStr.substring(len-3,len);
    Serial.println(cekStr);
   
   if(cekStr.equals("NO1")){
          digitalWrite(led1,HIGH); digitalWrite(led2,LOW); digitalWrite(led3,LOW); digitalWrite(buzzer,LOW);
          Serial.println("NO-1");
   } else if(cekStr.equals("NO2")){
          digitalWrite(led1,LOW); digitalWrite(led2,HIGH); digitalWrite(led3,LOW); digitalWrite(buzzer,LOW);
          Serial.println("NO-2");
   } else if(cekStr.equals("NO3")){
          digitalWrite(led1,HIGH); digitalWrite(led2,HIGH); digitalWrite(led3,LOW); digitalWrite(buzzer,LOW);
          Serial.println("NO-3");
   } else if(cekStr.equals("NO4")){
          digitalWrite(led1,LOW); digitalWrite(led2,LOW); digitalWrite(led3,HIGH); digitalWrite(buzzer,LOW);
          Serial.println("NO-4");
   } else if(cekStr.equals("NO5")){
          digitalWrite(led1,HIGH); digitalWrite(led2,LOW); digitalWrite(led3,HIGH); digitalWrite(buzzer,LOW);
          Serial.println("NO-5");
   } else if(cekStr.equals("NO6")){
          digitalWrite(led1,LOW); digitalWrite(led2,HIGH); digitalWrite(led3,HIGH); digitalWrite(buzzer,LOW);
          Serial.println("NO-6");
   } else if(cekStr.equals("NO7")){
          digitalWrite(led1,HIGH); digitalWrite(led2,HIGH); digitalWrite(led3,HIGH); digitalWrite(buzzer,LOW);
          Serial.println("NO-7");
   } else if(cekStr.equals("NO8")){
          digitalWrite(led1,HIGH); digitalWrite(led2,HIGH); digitalWrite(led3,HIGH); digitalWrite(buzzer,HIGH);
          Serial.println("NO-8");
   } else if(cekStr.equals("OKK")){
          digitalWrite(led1,LOW); digitalWrite(led2,LOW); digitalWrite(led3,LOW); digitalWrite(buzzer,LOW);
          Serial.println("OK-1");
   } else {
          digitalWrite(led1,LOW); digitalWrite(led2,LOW); digitalWrite(led3,LOW); digitalWrite(buzzer,HIGH);
          Serial.println("NO");
   }
    
  } else {
    Serial.println("connection failed");
  }
  
  client.stop();
  Serial.println();
  delay(3000);

}
