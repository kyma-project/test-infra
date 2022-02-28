import http from 'k6/http';
import { check } from 'k6';
var stringified = JSON.stringify(open('./body-istio.json'));
var kymaVersion = "2.0.4"
var requestCount = 1000

export default function () {
    var url = 'http://localhost:8080/v1/clusters'

    var now = Date.now();
    var clusterName= 'tst-'+now;


    var stringified1 = stringified.replace(/RUNTIME_ID/g, ++requestCount);
    stringified1 = stringified1.replace(/CLUSTER_NAME/g, clusterName);
    stringified1 = stringified1.replace(/KYMA_VERSION/g, kymaVersion)
    stringified1 = stringified1.replace(/GLOBAL_ACCOUNT_ID/g, now);
    stringified1 = stringified1.replace(/SERVICE_ID/g, now);
    stringified1 = stringified1.replace(/SERVICE_PLAN_ID/g, now);
    stringified1 = stringified1.replace(/SHOOT_NAME/g, now);
    stringified1 = stringified1.replace(/INSTANCE_ID/g, now);
    var payload = JSON.parse(stringified1);

    var params = {
        headers: {
            'Content-Type': 'application/json'
        },
    };
    let res = http.post(url, payload, params);
    var status = res.status;
    if( status != 200)
    {
        console.log("Response Code---->" + status)
        console.log("Response Body---->" + res.body)
        //errorRate.add(1)
    }

    check(res, { 'SuccessFull calls': (r) => r.status == 200 },);
}
