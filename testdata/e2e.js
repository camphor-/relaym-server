// ブラウザで /api/v3/login でログインした後、Chrome Dev Consoleに貼り付けてください。

console.log("----------STEP1 : 自分の情報を取得----------")

const getMeRes = await fetch("http://relaym.local:8080/api/v3/users/me", {
    "headers": {
        "X-CSRF-TOKEN": "relaym",
        "content-type":"application/json"
    },
    "method": "GET",
    "mode": "cors",
    "credentials": "include"
});
console.assert(getMeRes.ok,"自分の情報を取得に失敗しました",getMeRes.status)
const user = await getMeRes.json()
console.log(user)

console.log("----------STEP2 : デバイスを取得----------")

const getDevicesRes = await fetch("http://relaym.local:8080/api/v3/users/me/devices", {
    "headers": {
        "X-CSRF-TOKEN": "relaym",
        "content-type":"application/json"
    },
    "method": "GET",
    "mode": "cors",
    "credentials": "include"
});
console.assert(getDevicesRes.ok,"デバイスの取得に失敗しました",getDevicesRes.status)
const devices = await getDevicesRes.json()
console.log(devices)


console.log("----------STEP3 : セッションの作成----------")

const sessionName = 'test'
const createSessionRes = await fetch("http://relaym.local:8080/api/v3/sessions", {
    "headers": {
        "X-CSRF-TOKEN": "relaym",
        "content-type":"application/json"
    },
    "body": `{"name":"${sessionName}"}`,
    "method": "POST",
    "mode": "cors",
    "credentials": "include"
});
console.assert(createSessionRes.ok,"セッションの作成に失敗しました",createSessionRes.status)
const session = await createSessionRes.json()
console.log(session)

console.log("----------STEP4 : セッションに曲を追加(1曲目)----------")

const trackURI = 'spotify:track:5uQ0vKy2973Y9IUCd1wMEF'
const addQueueRes = await fetch(`http://relaym.local:8080/api/v3/sessions/${session.id}/queue`, {
    "headers": {
        "X-CSRF-TOKEN": "relaym",
        "content-type":"application/json"
    },
    "body": `{"uri":"${trackURI}"}`,
    "method": "POST",
    "mode": "cors",
    "credentials": "include"
});
console.assert(addQueueRes.ok,"キューへの追加に失敗しました",addQueueRes.status)

console.log("----------STEP5 : セッションに曲を追加(2曲目)----------")

const secondTrackURI = 'spotify:track:4BrmwVpDUCTNR0HzaO9Jpr'
const secondAddQueueRes = await fetch(`http://relaym.local:8080/api/v3/sessions/${session.id}/queue`, {
    "headers": {
        "X-CSRF-TOKEN": "relaym",
        "content-type":"application/json"
    },
    "body": `{"uri":"${secondTrackURI}"}`,
    "method": "POST",
    "mode": "cors",
    "credentials": "include"
});
console.assert(secondAddQueueRes.ok,"2曲目キューへの追加に失敗しました",secondAddQueueRes.status)

console.log("----------STEP6 : デバイスをセット----------")

const setDeviceRes = await fetch(`http://relaym.local:8080/api/v3/sessions/${session.id}/devices`, {
    "headers": {
        "X-CSRF-TOKEN": "relaym",
        "content-type":"application/json"
    },
    "body": `{"device_id":"${devices.devices[0].id}"}`,
    "method": "PUT",
    "mode": "cors",
    "credentials": "include"
});
console.assert(setDeviceRes.ok,"デバイスのセットに失敗しました",setDeviceRes.status)


console.log("----------STEP7 : 再生----------")

const playRes = await fetch(`http://relaym.local:8080/api/v3/sessions/${session.id}/playback`, {
    "headers": {
        "X-CSRF-TOKEN": "relaym",
        "content-type":"application/json"
    },
    "body": '{"state":"PLAY"}',
    "method": "PUT",
    "mode": "cors",
    "credentials": "include"
});
console.assert(playRes.ok,"曲の再生に失敗しました",playRes.status)

const sleep = msec => new Promise(resolve => setTimeout(resolve, msec))
await sleep(5000)

console.log("----------STEP8 : 一時停止----------")

const pauseRes = await fetch(`http://relaym.local:8080/api/v3/sessions/${session.id}/playback`, {
    "headers": {
        "X-CSRF-TOKEN": "relaym",
        "content-type":"application/json"
    },
    "body": '{"state":"PAUSE"}',
    "method": "PUT",
    "mode": "cors",
    "credentials": "include"
});
console.assert(pauseRes.ok,"曲の一時停止に失敗しました",pauseRes.status)
await sleep(5000)

console.log("----------STEP9 : 再度再生----------")

const rePlayRes = await fetch(`http://relaym.local:8080/api/v3/sessions/${session.id}/playback`, {
    "headers": {
        "X-CSRF-TOKEN": "relaym",
        "content-type":"application/json"
    },
    "body": '{"state":"PLAY"}',
    "method": "PUT",
    "mode": "cors",
    "credentials": "include"
});
console.assert(rePlayRes.ok,"曲の再度再生に失敗しました",rePlayRes.status)

console.log("----------STEP10 : セッションに曲を追加(三曲目)----------")

const thirdTrackURI = 'spotify:track:36MwbWxhiFdyD29fhWJUoQ'
const thirdAddQueueRes = await fetch(`http://relaym.local:8080/api/v3/sessions/${session.id}/queue`, {
    "headers": {
        "X-CSRF-TOKEN": "relaym",
        "content-type":"application/json"
    },
    "body": `{"uri":"${thirdTrackURI}"}`,
    "method": "POST",
    "mode": "cors",
    "credentials": "include"
});
console.assert(thirdAddQueueRes.ok,"3曲目のキューへの追加に失敗しました",thirdAddQueueRes.status)