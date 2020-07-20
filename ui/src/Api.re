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

let p = "http://localhost:8080/api";

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

let fetchJSONResponse = res => {
  Fetch.Response.ok(res)
    ? Fetch.Response.json(res)
    : Js.Exn.raiseError(Fetch.Response.statusText(res));
};
let fetchTextResponse = res => {
  Fetch.Response.ok(res)
    ? Fetch.Response.text(res)
    : Js.Exn.raiseError(Fetch.Response.statusText(res));
};

let postTest = (payload, callback) => {
  let path = p ++ "/tests";
  Js.Promise.(
    JsonTransport.post(path, payload)
    |> then_(res => res |> fetchJSONResponse)
    |> then_(response => callback(SuccessJSON(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let putTest = (payload, callback) => {
  let path = p ++ "/tests";
  Js.Promise.(
    JsonTransport.put(path, payload)
    |> then_(res => res |> fetchJSONResponse)
    |> then_(response => callback(SuccessJSON(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let deleteTest = (id, callback) => {
  let path = p ++ "/tests/" ++ id;
  Js.Promise.(
    JsonTransport.delete(path)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let tryTest = (payload, callback) => {
  let path = p ++ "/tests/test";
  Js.Promise.(
    JsonTransport.post(path, payload)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let postContact = (payload, callback) => {
  let path = p ++ "/contacts";
  Js.Promise.(
    JsonTransport.post(path, payload)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let putContact = (payload, callback) => {
  let path = p ++ "/contacts";
  Js.Promise.(
    JsonTransport.put(path, payload)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let deleteContact = (id, callback) => {
  let path = p ++ "/contacts/" ++ id;
  Js.Promise.(
    JsonTransport.delete(path)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let tryContact = (payload, callback) => {
  let path = p ++ "/contacts/test";
  Js.Promise.(
    JsonTransport.post(path, payload)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let postTestContacts = (payload, callback) => {
  let path = p ++ "/testcontacts";
  Js.Promise.(
    JsonTransport.post(path, payload)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};

let putTestContacts = (payload, callback) => {
  let path = p ++ "/testcontacts";
  Js.Promise.(
    JsonTransport.put(path, payload)
    |> then_(res => res |> fetchTextResponse)
    |> then_(response => callback(Success(response)) |> resolve)
    |> catch(e => callback(Error([%bs.raw "e.message"])) |> resolve)
    |> ignore
  );
};