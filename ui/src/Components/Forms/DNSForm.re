type dnsFormTypes =
  /* Poll tests */
  | TestName
  | Interval
  | Timeout
  | Url
  /* Contacts */
  | Contacts
  /* DNS */
  | Record
  | Strategy
  | Check;

type dnsFormState = {
  /* Poll tests */
  mutable testName: Form.t,
  mutable interval: Form.t,
  mutable timeout: Form.t,
  mutable url: Form.t,
  /* Contacts */
  mutable contacts: Form.t,
  /* DNS */
  mutable record: Form.t,
  mutable strategy: Form.t,
  mutable check: Form.t,
};

let getInitialFormState = () => {
  testName: Str(""),
  interval: Int(0),
  timeout: Int(0),
  url: Str(""),
  contacts: TupleList([]),
  record: Str(""),
  strategy: Str(""),
  check: List([""]),
};

module DNSFormConfig = {
  type field = dnsFormTypes;
  type state = dnsFormState;
  let update = (field, value, state) => {
    switch (field, value) {
    | (TestName, v) => {...state, testName: v}
    | (Interval, v) => {...state, interval: v}
    | (Timeout, v) => {...state, timeout: v}
    | (Url, v) => {...state, url: v}
    | (Contacts, v) => {...state, contacts: v}
    | (Record, v) => {...state, record: v}
    | (Strategy, v) => {...state, strategy: v}
    | (Check, v) => {...state, check: v}
    };
  };
  let get = (field, state) => {
    switch (field) {
    | TestName => state.testName
    | Interval => state.interval
    | Timeout => state.timeout
    | Url => state.url
    | Contacts => state.contacts
    | Record => state.record
    | Strategy => state.strategy
    | Check => state.check
    };
  };
};

module DNSForm = Form.FormComponent(DNSFormConfig);

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
  (Record, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
  (Strategy, [(Form.NotEmpty, FormHelpers.emptyMsg)]),
  (
    Check,
    [
      (Form.NotEmpty, FormHelpers.emptyMsg),
      (
        Form.Custom(FormHelpers.listValidation),
        "Fill all values or remove unused once",
      ),
    ],
  ),
];

let getTestPayload = (~inputTest=?, values) => {
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

  FormHelpers.setJsonKey(payload, "test_type", Str("DNS"));
  FormHelpers.setJsonKey(payload, "test_name", values.testName);
  FormHelpers.setJsonKey(payload, "timeout", values.timeout);
  FormHelpers.setJsonKey(payload, "url", values.url);
  FormHelpers.setJsonKey(payload, "interval", values.interval);

  let blob = Js.Dict.empty();
  FormHelpers.setJsonKey(blob, "record", values.record);
  FormHelpers.setJsonKey(blob, "strategy", values.strategy);
  FormHelpers.setJsonKey(blob, "check", values.check);
  Js.Dict.set(payload, "blob", Js.Json.object_(blob));

  payload;
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
  <DNSForm
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
                     | DNS(dns) =>
                       init.record = Str(dns.record);
                       init.strategy = Str(dns.strategy);
                       init.check = List(dns.check);
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
    render={(f: DNSForm.form) =>
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
            infoText="Hostname which DNS will be checked (*)"
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
          <FormSelect
            width=Half
            label="Record type"
            placeholder="DNS record"
            options=[
              ("Address", "A"),
              ("CNAME", "CNAME"),
              ("TXT", "TXT"),
              ("MX", "MX"),
              ("NS", "NS"),
            ]
            value={f.form.values.record}
            onChange={v => v |> f.handleChange(Record)}
            infoText="The record type which will be checked (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Record, f.form.errors) : React.null
            }
          />
          <FormSelect
            width=Half
            label="Check strategy"
            placeholder="Check strategy"
            options=[
              ("CheckIsSubset", "check_is_subset"),
              ("DNSIsSubset", "dns_is_subset"),
              ("Intersects", "intersects"),
              ("Exact", "exact"),
            ]
            value={f.form.values.strategy}
            onChange={v => v |> f.handleChange(Strategy)}
            infoText="The way in which record and DNS values will be compared (*)"
            errorMsg={
              submitted
                ? FormHelpers.getError(Strategy, f.form.errors) : React.null
            }
          />
        </div>
        <div className="flex flex-wrap -mx-3 mb-6">
          <FormList
            label="check values"
            errorMsg={
              submitted
                ? FormHelpers.getError(Check, f.form.errors) : React.null
            }
            list={
              switch (f.form.values.check) {
              | List(l) => l
              | _ => []
              }
            }
            placeholder="192.168.0.1"
            onChange={v => v |> f.handleChange(Check)}
            infoText="Values which will be compared with the DNS lookup (*)"
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