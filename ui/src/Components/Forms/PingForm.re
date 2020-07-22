type pingFormTypes =
  /* Poll tests */
  | TestName
  | Interval
  | Timeout
  | Url;

type pingFormState = {
  /* Poll tests */
  mutable testName: Form.t,
  mutable interval: Form.t,
  mutable timeout: Form.t,
  mutable url: Form.t,
};

let getInitialFormState = () => {
  testName: Str(""),
  interval: Int(0),
  timeout: Int(0),
  url: Str(""),
};

module PingFormConfig = {
  type field = pingFormTypes;
  type state = pingFormState;
  let update = (field, value, state) => {
    switch (field, value) {
    | (TestName, v) => {...state, testName: v}
    | (Interval, v) => {...state, interval: v}
    | (Timeout, v) => {...state, timeout: v}
    | (Url, v) => {...state, url: v}
    };
  };
  let get = (field, state) => {
    switch (field) {
    | TestName => state.testName
    | Interval => state.interval
    | Timeout => state.interval
    | Url => state.url
    };
  };
};

module PingForm = Form.FormComponent(PingFormConfig);

let rules = [
  (TestName, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
  (Interval, [(Form.NotEmpty, FormHelpers.aboveZero)]),
  (Timeout, [(Form.NotEmpty, FormHelpers.aboveZero)]),
  (Url, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
];

let getPayload = (~inputTest=?, values) => {
  let payload = Js.Dict.empty();
  switch (inputTest) {
  | Some(test) =>
    Models.Test.(
      switch (test) {
      | Some(test_) =>
        FormHelpers.setJsonKey(payload, "test_id", Str(test_.testId))
      | None => ()
      }
    )

  | None => ()
  };

  FormHelpers.setJsonKey(payload, "test_type", Str("Ping"));
  FormHelpers.setJsonKey(payload, "test_name", values.testName);
  FormHelpers.setJsonKey(payload, "timeout", values.timeout);
  FormHelpers.setJsonKey(payload, "url", values.url);
  FormHelpers.setJsonKey(payload, "interval", values.interval);

  let blob = Js.Dict.empty();
  Js.Dict.set(payload, "blob", Js.Json.object_(blob));

  payload;
};

[@react.component]
let make =
    (
      ~submitTest,
      /*~submitContacts,*/
      ~inputTest: option(Models.Test.t)=?,
      ~inputTestContacts=?,
    ) => {
  let (submitted, setSubmitted) = React.useState(_ => false);
  let (submitError, setSubmitError) = React.useState(() => "");
  let (tryTestMsg, setTryTestMsg) = React.useState(() => "");
  let (testContacts, setTestContacts) =
    React.useState(_ =>
      switch (inputTestContacts) {
      | Some(testContacts) => testContacts
      | None => []
      }
    );

  let postCallback = resp => {
    switch (resp) {
    | Api.Error(msg) => setSubmitError(_ => msg)
    | Api.SuccessJSON(_) =>
      /*if (List.length(testContacts) > 0) {
          submitContacts(
            testContactPayloadOfState(
              Models.Decode.test(jsonTest).testId,
              testContacts,
            ),
            postCallback,
          );
        } else {*/
      Paths.goToTests()
    /*}*/
    | Api.Success(_msg) => Paths.goToTests()
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
      let payload = getPayload(values, ~inputTest);
      submitTest(payload, postCallback);
    };
  };

  let handleTryTest = (values, errors) => {
    setSubmitted(_ => true);
    if (List.length(errors) == 0) {
      setTryTestMsg(_ => "Loading...");
      let payload = getPayload(values);
      Api.tryTest(payload, tryTestCallback);
    };
  };
  <PingForm
    initialState={
                   let init = getInitialFormState();
                   switch (inputTest) {
                   | None => ()
                   | Some(test) =>
                     init.testName = Str(test.testName);
                     init.interval = Int(test.interval);
                     init.timeout = Int(test.timeout);
                     init.url = Str(test.url);
                   };
                   init;
                 }
    rules
    render={(f: PingForm.form) =>
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
            infoText="Hostname that will be pinged (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Url, f.form.errors) : React.null
            }
            placeholder="google.com"
            value={f.form.values.url}
            onChange={v => v |> f.handleChange(Url)}
          />
        </div>
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