type portFormTypes =
  /* Poll tests */
  | TestName
  | Interval
  | Timeout
  | Url
  /* Contacts */
  | Contacts
  /* Port tests */
  | Port;

type portFormState = {
  /* Poll tests */
  mutable testName: Form.t,
  mutable interval: Form.t,
  mutable timeout: Form.t,
  mutable url: Form.t,
  /* Contacts */
  mutable contacts: Form.t,
  /* Port tests */
  mutable port: Form.t,
};

let getInitialFormState = () => {
  testName: Str(""),
  interval: Int(0),
  timeout: Int(0),
  url: Str(""),
  contacts: TupleList([]),
  port: Str(""),
};

module PortFormConfig = {
  type field = portFormTypes;
  type state = portFormState;
  let update = (field, value, state) => {
    switch (field, value) {
    | (TestName, v) => {...state, testName: v}
    | (Interval, v) => {...state, interval: v}
    | (Timeout, v) => {...state, timeout: v}
    | (Url, v) => {...state, url: v}
    | (Contacts, v) => {...state, contacts: v}
    | (Port, v) => {...state, port: v}
    };
  };
  let get = (field, state) => {
    switch (field) {
    | TestName => state.testName
    | Interval => state.interval
    | Timeout => state.timeout
    | Url => state.url
    | Contacts => state.contacts
    | Port => state.port
    };
  };
};

module PortForm = Form.FormComponent(PortFormConfig);

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
  (Port, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
];

let getTestPayload = (~inputTest=?, values, testType) => {
  let payload = Js.Dict.empty();
  switch (inputTest) {
  | Some(test) =>
    Models.Test.(
      switch (test) {
      | Some(test_) =>
        FormHelpers.setJsonKey(payload, "test_id", Str(test_.testId));
        Js.Dict.set(payload, "active", Js.Json.boolean(test_.active));
      | None => Js.Dict.set(payload, "active", Js.Json.boolean(true))
      }
    )

  | None => ()
  };

  FormHelpers.setJsonKey(payload, "test_type", Str(testType));
  FormHelpers.setJsonKey(payload, "test_name", values.testName);
  FormHelpers.setJsonKey(payload, "timeout", values.timeout);
  FormHelpers.setJsonKey(payload, "url", values.url);
  FormHelpers.setJsonKey(payload, "interval", values.interval);

  let blob = Js.Dict.empty();
  FormHelpers.setJsonKey(blob, "port", values.port);
  Js.Dict.set(payload, "blob", Js.Json.object_(blob));

  payload;
};

[@react.component]
let make =
    (
      ~testType,
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
      let payload = getTestPayload(values, testType, ~inputTest);
      submitTest(payload, postCallback(values));
    };
  };

  let handleTryTest = (values, errors) => {
    setSubmitted(_ => true);
    if (List.length(errors) == 0) {
      setTryTestMsg(_ => "Loading...");
      let payload = getTestPayload(values, testType);
      Api.tryTest(payload, tryTestCallback);
    };
  };
  <PortForm
    initialState={
                   let init = getInitialFormState();
                   switch (inputTest) {
                   | None => ()
                   | Some(test) =>
                     init.testName = Str(test.testName);
                     init.interval = Int(test.interval);
                     init.timeout = Int(test.timeout);
                     init.url = Str(test.url);
                     switch (test.specific) {
                     | TCP(port) => init.port = Str(port.port)
                     | _ => ()
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
    render={(f: PortForm.form) =>
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
            label="Hostname"
            infoText="Hostname that will be polled (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Url, f.form.errors) : React.null
            }
            placeholder="google.com"
            value={f.form.values.url}
            onChange={v => v |> f.handleChange(Url)}
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormInput
            type_=Text
            width=Full
            label="Port"
            infoText="The port of the hostname (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Port, f.form.errors) : React.null
            }
            placeholder="443"
            value={f.form.values.port}
            onChange={v => v |> f.handleChange(Port)}
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