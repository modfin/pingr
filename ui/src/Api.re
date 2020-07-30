type response =
  | Success(string)
  | SuccessJSON(Js.Json.t)
  | Error(string);

module JsonTransport = {
  let headers = Fetch.HeadersInit.make({"Content-Type": "application/json"});

  let get = (~path) => {
    Js.Promise.(Fetch.fetch(path) |> then_(Fetch.Response.json));
  };

  let post = (path: string, payload) => {
    let body_ = Js.Json.stringifyAny(payload);

    switch (body_) {
    | Some(body) =>
      Fetch.fetchWithInit(
        path,
        Fetch.RequestInit.make(
          ~method_=Post,
          ~body=Bs_fetch.BodyInit.make(body),
          ~headers,
          (),
        ),
      )
    | None =>
      Fetch.fetchWithInit(
        path,
        Fetch.RequestInit.make(~method_=Post, ~headers, ()),
      )
    };
  };

  let put = (path: string, payload) => {
    let body_ = Js.Json.stringifyAny(payload);
    switch (body_) {
    | Some(body) =>
      Fetch.fetchWithInit(
        path,
        Fetch.RequestInit.make(
          ~method_=Put,
          ~body=Bs_fetch.BodyInit.make(body),
          ~headers,
          (),
        ),
      )
    | None =>
      Fetch.fetchWithInit(
        path,
        Fetch.RequestInit.make(~method_=Put, ~headers, ()),
      )
    };
  };

  let delete = path => {
    Fetch.fetchWithInit(path, Fetch.RequestInit.make(~method_=Delete, ()));
  };
};

let p = "/api";

let fetchJson = (url, decoder) =>
  Js.Promise.(
    Fetch.fetch(url)
    |> then_(Fetch.Response.json)
    |> then_(obj => obj |> decoder |> resolve)
  );

let fetchJsonWithCallback = (url, decoder, callback) =>
  Js.Promise.(
    fetchJson(url, decoder)
    |> then_(result => callback(Some(result)) |> resolve)
    |> catch(_err => callback(None) |> resolve)
    |> ignore
  );

let fetchTestsWithCallback = callback => {
  fetchJsonWithCallback(p ++ "/tests", Models.Decode.tests, callback);
};

let fetchTestWithCallback = (id, callback) => {
  fetchJsonWithCallback(p ++ "/tests/" ++ id, Models.Decode.test, callback);
};

let fetchTestsStatusWithCallback = callback => {
  fetchJsonWithCallback(
    p ++ "/tests/status",
    Models.Decode.testsStatus,
    callback,
  );
};

let fetchLogsWithCallback = (~id=?, callback, ~numDays=?, ()) => {
  let path =
    p
    ++ (
      switch (id, numDays) {
      | (None, _) => "/logs"
      | (Some(id), None)
      | (Some(id), Some("0")) => "/tests/" ++ id ++ "/logs"
      | (Some(id), Some(days)) => "/tests/" ++ id ++ "/logs/" ++ days
      }
    );
  fetchJsonWithCallback(path, Models.Decode.logs, callback);
};

let fetchContactsWithCallback = callback => {
  fetchJsonWithCallback(p ++ "/contacts", Models.Decode.contacts, callback);
};

let fetchContactWithCallback = (id, callback) => {
  fetchJsonWithCallback(
    p ++ "/contacts/" ++ id,
    Models.Decode.contact,
    callback,
  );
};

let fetchTestContactsWithCallback = (id, callback) => {
  fetchJsonWithCallback(
    p ++ "/testcontacts/" ++ id ++ "/types",
    Models.Decode.testContacts,
    callback,
  );
};

let fetchIncidentWithCallback = (id, callback) => {
  fetchJsonWithCallback(
    p ++ "/incidents/" ++ id,
    Models.Decode.incidentFull,
    callback,
  );
};

let fetchIncidentsWithCallback = callback => {
  fetchJsonWithCallback(p ++ "/incidents", Models.Decode.incidents, callback);
};

let fetchJSONResponse = res => {
  Fetch.Response.ok(res)
    ? Fetch.Response.json(res)
    : Js.Exn.raiseError(Fetch.Response.statusText(res));
};
let fetchTextResponse = res => {
  Fetch.Response.ok(res)
    ? Fetch.Response.text(res)
    : Fetch.Response.text(res)
      |> Js.Promise.then_(res => Js.Exn.raiseError(res) |> Js.Promise.reject);
};

let postWithJsonResponse = (path, payload, callback) => {
  Js.Promise.(
    JsonTransport.post(path, payload)
    |> then_(res => res |> fetchJSONResponse)
    |> then_(response => callback(SuccessJSON(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let postWithTextResponse = (path, payload, callback) => {
  Js.Promise.(
    JsonTransport.post(path, payload)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let putWithJsonResponse = (path, payload, callback) => {
  Js.Promise.(
    JsonTransport.put(path, payload)
    |> then_(res => res |> fetchJSONResponse)
    |> then_(response => callback(SuccessJSON(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let putWithTextResponse = (path, payload, callback) => {
  Js.Promise.(
    JsonTransport.put(path, payload)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let deleteWithTextResponse = (path, callback) => {
  Js.Promise.(
    JsonTransport.delete(path)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let postTest = (payload, callback) => {
  let path = p ++ "/tests";
  postWithJsonResponse(path, payload, callback);
};

let putTest = (payload, callback) => {
  let path = p ++ "/tests";
  putWithJsonResponse(path, payload, callback);
};

let deleteTest = (id, callback) => {
  let path = p ++ "/tests/" ++ id;
  deleteWithTextResponse(path, callback);
};

let tryTest = (payload, callback) => {
  let path = p ++ "/tests/test";
  postWithTextResponse(path, payload, callback);
};

let updateActive = (testId, value, callback) => {
  let path =
    p
    ++ "/tests/"
    ++ testId
    ++ "/"
    ++ {
      value ? "activate" : "deactivate";
    };
  putWithTextResponse(path, Js.Dict.empty(), callback);
};

let postContact = (payload, callback) => {
  let path = p ++ "/contacts";
  postWithTextResponse(path, payload, callback);
};

let putContact = (payload, callback) => {
  let path = p ++ "/contacts";
  putWithTextResponse(path, payload, callback);
};

let deleteContact = (id, callback) => {
  let path = p ++ "/contacts/" ++ id;
  deleteWithTextResponse(path, callback);
};

let tryContact = (payload, callback) => {
  let path = p ++ "/contacts/test";
  postWithTextResponse(path, payload, callback);
};

let postTestContacts = (payload, callback) => {
  let path = p ++ "/testcontacts";
  postWithTextResponse(path, payload, callback);
};

let putTestContacts = (payload, callback) => {
  let path = p ++ "/testcontacts";
  putWithTextResponse(path, payload, callback);
};