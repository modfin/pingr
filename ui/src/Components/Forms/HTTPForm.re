type httpFormTypes =
  /* Poll tests */
  | TestName
  | Interval
  | Timeout
  | Url
  /* Contacts */
  | Contacts
  /* Http tests */
  | ReqMethod
  | ReqHeaders
  | ReqBody
  | ResStatus
  | ResHeaders
  | ResBody;

type httpFormState = {
  /* Poll tests */
  mutable testName: Form.t,
  mutable interval: Form.t,
  mutable timeout: Form.t,
  mutable url: Form.t,
  /* Contacts */
  mutable contacts: Form.t,
  /* Http tests */
  mutable reqMethod: Form.t,
  mutable reqHeaders: Form.t,
  mutable reqBody: Form.t,
  mutable resStatus: Form.t,
  mutable resHeaders: Form.t,
  mutable resBody: Form.t,
};

let getInitialFormState = () => {
  testName: Str(""),
  interval: Int(0),
  timeout: Int(0),
  url: Str(""),
  contacts: TupleList([]),
  reqMethod: Str(""),
  reqHeaders: TupleList([]),
  reqBody: Str(""),
  resStatus: Int(0),
  resHeaders: TupleList([]),
  resBody: Str(""),
};

module HttpFormConfig = {
  type field = httpFormTypes;
  type state = httpFormState;
  let update = (field, value, state) => {
    switch (field, value) {
    | (TestName, v) => {...state, testName: v}
    | (Interval, v) => {...state, interval: v}
    | (Timeout, v) => {...state, timeout: v}
    | (Url, v) => {...state, url: v}
    | (Contacts, v) => {...state, contacts: v}
    | (ReqMethod, v) => {...state, reqMethod: v}
    | (ReqHeaders, v) => {...state, reqHeaders: v}
    | (ReqBody, v) => {...state, reqBody: v}
    | (ResStatus, v) => {...state, resStatus: v}
    | (ResHeaders, v) => {...state, resHeaders: v}
    | (ResBody, v) => {...state, resBody: v}
    };
  };
  let get = (field, state) => {
    switch (field) {
    | TestName => state.testName
    | Interval => state.interval
    | Timeout => state.timeout
    | Url => state.url
    | Contacts => state.contacts
    | ReqMethod => state.reqMethod
    | ReqHeaders => state.reqHeaders
    | ReqBody => state.reqBody
    | ResStatus => state.resStatus
    | ResHeaders => state.resHeaders
    | ResBody => state.resBody
    };
  };
};

module HttpForm = Form.FormComponent(HttpFormConfig);
let keyMsg = "Fill keys and values or remove unused one's";

let rules = [
  (TestName, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
  (Interval, [(Form.NotEmpty, FormHelpers.aboveZero)]),
  (Timeout, [(Form.NotEmpty, FormHelpers.aboveZero)]),
  (Url, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
  (
    Contacts,
    [
      (
        Form.Custom(FormHelpers.validContactThreshold),
        FormHelpers.thresholdMsg,
      ),
    ],
  ),
  (ReqMethod, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
  (ReqHeaders, [(Form.Custom(FormHelpers.keyPairValidation), keyMsg)]),
  (ResHeaders, [(Form.Custom(FormHelpers.keyPairValidation), keyMsg)]),
];

let httpMethods = [
  ("GET", "GET"),
  ("POST", "POST"),
  ("PUT", "PUT"),
  ("HEAD", "HEAD"),
  ("DELETE", "DELETE"),
];

let getTestPayload = (~inputTest=?, values) => {
  let testPayload = Js.Dict.empty();
  switch (inputTest) {
  | Some(test) =>
    Models.Test.(
      switch (test) {
      | Some(test_) =>
        FormHelpers.setJsonKey(testPayload, "test_id", Str(test_.testId));
        Js.Dict.set(testPayload, "active", Js.Json.boolean(test_.active));
      | None => Js.Dict.set(testPayload, "active", Js.Json.boolean(true))
      }
    )

  | None => ()
  };

  FormHelpers.setJsonKey(testPayload, "test_type", Str("HTTP"));
  FormHelpers.setJsonKey(testPayload, "test_name", values.testName);
  FormHelpers.setJsonKey(testPayload, "timeout", values.timeout);
  FormHelpers.setJsonKey(testPayload, "url", values.url);
  FormHelpers.setJsonKey(testPayload, "interval", values.interval);

  let blob = Js.Dict.empty();
  FormHelpers.setJsonKey(blob, "req_method", values.reqMethod);
  if (values.reqHeaders == TupleList([])) {
    Js.Dict.set(blob, "req_headers", Js.Json.object_(Js.Dict.empty()));
  } else {
    FormHelpers.setJsonKey(blob, "req_headers", values.reqHeaders);
  };
  FormHelpers.setJsonKey(blob, "req_body", values.reqBody);
  FormHelpers.setJsonKey(blob, "res_status", values.resStatus);
  if (values.resHeaders == TupleList([])) {
    Js.Dict.set(blob, "res_headers", Js.Json.object_(Js.Dict.empty()));
  } else {
    FormHelpers.setJsonKey(blob, "res_headers", values.resHeaders);
  };
  FormHelpers.setJsonKey(blob, "res_body", values.resBody);
  Js.Dict.set(testPayload, "blob", Js.Json.object_(blob));

  testPayload;
};

[@react.component]
let make =
    (
      ~submitTest,
      ~submitContacts,
      ~inputTest: option(Models.Test.t)=?,
      ~inputTestContacts=?,
    ) => {
  let (submitted, setSubmitted) = React.useState(_ => false);
  let (submitError, setSubmitError) = React.useState(() => "");
  let (tryTestMsg, setTryTestMsg) = React.useState(() => "");

  let rec postCallback = (values, resp) => {
    switch (resp) {
    | Api.Error(msg) => setSubmitError(_ => msg)
    | Api.SuccessJSON(testJson) =>
      let testContacts =
        switch (values.contacts) {
        | TupleList(l) => l
        | _ => []
        };
      if (List.length(testContacts) > 0) {
        let id = Models.Decode.test(testJson).testId;
        submitContacts(
          FormHelpers.getContactsPayload(id, TupleList(testContacts)),
          postCallback(values),
        );
      } else {
        Paths.goToTests();
      };
    | Success(_) => Paths.goToTests()
    };
  };

  let tryTestCallback = resp => {
    switch (resp) {
    | Api.Error(msg) =>
      setSubmitError(_ => msg);
      setTryTestMsg(_ => "");
    | Api.Success(msg) =>
      setTryTestMsg(_ => msg);
      setSubmitError(_ => "");
    | Api.SuccessJSON(_json) => ()
    };
  };

  let handleSubmit = (e, values, errors) => {
    ReactEvent.Form.preventDefault(e);
    setSubmitted(_ => true);
    if (List.length(errors) == 0) {
      let payload = getTestPayload(values, ~inputTest);
      submitTest(payload, postCallback(values));
    };
  };

  let handleTryTest = (values, errors) => {
    setSubmitted(_ => true);
    if (List.length(errors) == 0) {
      setTryTestMsg(_ => "Loading...");
      let payload = getTestPayload(values);
      Api.tryTest(payload, tryTestCallback);
    };
  };

  <HttpForm
    initialState={
                   let init = getInitialFormState();
                   switch (inputTest) {
                   | None => ()
                   | Some(httpTest) =>
                     init.testName = Str(httpTest.testName);
                     init.interval = Int(httpTest.interval);
                     init.timeout = Int(httpTest.timeout);
                     init.url = Str(httpTest.url);
                     switch (httpTest.specific) {
                     | HTTP(http) =>
                       init.reqMethod = Str(http.reqMethod);
                       switch (http.reqHeaders) {
                       | None => ()
                       | Some(headers) =>
                         init.reqHeaders =
                           TupleList(
                             Js.Dict.entries(headers) |> Array.to_list,
                           )
                       };
                       switch (http.reqBody) {
                       | None => ()
                       | Some(body) => init.reqBody = Str(body)
                       };
                       switch (http.resStatus) {
                       | None => ()
                       | Some(status) => init.resStatus = Int(status)
                       };
                       switch (http.resHeaders) {
                       | None => ()
                       | Some(headers) =>
                         init.resHeaders =
                           TupleList(
                             Js.Dict.entries(headers) |> Array.to_list,
                           )
                       };
                       switch (http.resBody) {
                       | None => ()
                       | Some(body) => init.resBody = Str(body)
                       };
                     | _ => () /* Should only be http test*/
                     };
                   };
                   switch (inputTestContacts) {
                   | None => ()
                   | Some(testContacts) =>
                     init.contacts =
                       TupleList(
                         testContacts
                         |> List.map(testContact => {
                              Models.TestContact.(
                                testContact.contactId,
                                string_of_int(testContact.threshold),
                              )
                            }),
                       )
                   };
                   init;
                 }
    rules
    render={(f: HttpForm.form) =>
      <form
        onSubmit={e => handleSubmit(e, f.form.values, f.form.errors)}
        className="w-full">
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormInput
            type_=Text
            width=Full
            label="Name"
            placeholder="Some name"
            infoText="Test's name, no functional meaning (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(TestName, f.form.errors) : React.null
            }
            value={f.form.values.testName}
            onChange={v => v |> f.handleChange(TestName)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormInput
            type_=Number
            width=Half
            label="Interval (s)"
            infoText="The number of seconds between each test (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Interval, f.form.errors) : React.null
            }
            value={f.form.values.interval}
            onChange={v => v |> f.handleChange(Interval)}
          />
          <FormInput
            type_=Number
            width=Half
            label="Timeout (s)"
            infoText="The number of seconds before test times out (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Timeout, f.form.errors) : React.null
            }
            value={f.form.values.timeout}
            onChange={v => v |> f.handleChange(Timeout)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormInput
            type_=Text
            width=Full
            label="Url"
            infoText="Url that test will poll against (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Url, f.form.errors) : React.null
            }
            placeholder="https://google.com"
            value={f.form.values.url}
            onChange={v => v |> f.handleChange(Url)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormSelect
            width=Half
            label="http method"
            placeholder="Choose HTTP method"
            infoText="HTTP method (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(ReqMethod, f.form.errors) : React.null
            }
            options=httpMethods
            value={f.form.values.reqMethod}
            onChange={v => v |> f.handleChange(ReqMethod)}
          />
          <FormInput
            type_=Number
            width=Half
            label="Accepted response code"
            infoText="Expected response HTTP status code"
            errorMsg={
              submitted
                ? FormHelpers.getError(ResStatus, f.form.errors) : React.null
            }
            value={f.form.values.resStatus}
            onChange={v => v |> f.handleChange(ResStatus)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormKeyValue
            pairs={
              switch (f.form.values.reqHeaders) {
              | TupleList(l) => l
              | _ => []
              }
            }
            label="request headers"
            infoText="Values for the request header"
            keyPlaceholder="Content-Type"
            valuePlaceholder="application/json"
            errorMsg={
              submitted
                ? FormHelpers.getError(ReqHeaders, f.form.errors) : React.null
            }
            onChange={v => v |> f.handleChange(ReqHeaders)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormTextarea
            label="Request body"
            placeholder="{ token: 1234-abcd, ...}"
            infoText="Body of the request"
            errorMsg={
              submitted
                ? FormHelpers.getError(ReqBody, f.form.errors) : React.null
            }
            value={f.form.values.reqBody}
            onChange={v => v |> f.handleChange(ReqBody)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormKeyValue
            pairs={
              switch (f.form.values.resHeaders) {
              | TupleList(l) => l
              | _ => []
              }
            }
            label="response headers"
            infoText="Values that are expected in the response header"
            keyPlaceholder="Content-Type"
            valuePlaceholder="application/json"
            errorMsg={
              submitted
                ? FormHelpers.getError(ResHeaders, f.form.errors) : React.null
            }
            onChange={v => v |> f.handleChange(ResHeaders)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormTextarea
            label="Expected response body"
            placeholder="pong"
            infoText="Expected response body"
            errorMsg={
              submitted
                ? FormHelpers.getError(ResBody, f.form.errors) : React.null
            }
            value={f.form.values.resBody}
            onChange={v => v |> f.handleChange(ResBody)}
          />
        </div>
        <FormTestContacts
          errorMsg={
            submitted
              ? FormHelpers.getError(Contacts, f.form.errors) : React.null
          }
          value={
            switch (f.form.values.contacts) {
            | TupleList(l) => l
            | _ => []
            }
          }
          onChange={v => v |> f.handleChange(Contacts)}
        />
        <button
          type_="button"
          onClick={_ => handleTryTest(f.form.values, f.form.errors)}
          className="mr-1 bg-blue-500 hover:bg-blue-700 text-white py-1 px-2 rounded">
          {"Try test" |> React.string}
        </button>
        <button
          type_="submit"
          className="m-1 bg-green-500 hover:bg-green-700 text-white py-1 px-2 rounded">
          {"Submit" |> React.string}
        </button>
        {tryTestMsg != ""
           ? <p className="text-gray-600 mb-2">
               {tryTestMsg |> React.string}
             </p>
           : React.null}
        {submitError != ""
           ? <p className="text-red-500">
               {"Error posting test: " ++ submitError |> React.string}
             </p>
           : React.null}
        <div className="h-32" />
      </form>
    }
  />;
};