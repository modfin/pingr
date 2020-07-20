module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadData
  | LoadSuccess(list(Models.Contact.t))
  | LoadFail(string);

type formTypes =
  | TestType /* Poll Push */
  | TestName /* Poll Push */
  | Url /* Poll */
  | Interval /* Poll */
  | Timeout /* Poll Push */
  | Port /* SSH TCP TLS */
  | Method /* HTTP */
  | Payload /* HTTP */
  | ExpResult /* HTTP */
  | Username /* SSH */
  | Password /* SSH */
  | UseKeyPair /* SSH */
  | Key /* Prometheus */
  | LowerBound /* Prometheus */
  | UpperBound /* Prometheus */
  | Labels /* Prometheus */
  | A /* DNS */
  | CNAME /* DNS */
  | TXT; /* DNS */

type formState = {
  mutable testType: Form.t,
  mutable testName: Form.t,
  mutable url: Form.t,
  mutable interval: Form.t,
  mutable timeout: Form.t,
  mutable port: Form.t,
  mutable method_: Form.t,
  mutable payload: Form.t,
  mutable expResult: Form.t,
  mutable username: Form.t,
  mutable password: Form.t,
  mutable useKeyPair: Form.t,
  mutable key: Form.t,
  mutable lowerBound: Form.t,
  mutable upperBound: Form.t,
  mutable labels: Form.t,
  mutable a: Form.t,
  mutable cname: Form.t,
  mutable txt: Form.t,
};

let getInitialFormState = () => {
  {
    testType: Str(""),
    testName: Str(""),
    url: Str(""),
    interval: Int(0),
    timeout: Int(0),
    port: Str(""),
    method_: Str(""),
    payload: Str(""),
    expResult: Str(""),
    username: Str(""),
    password: Str(""),
    useKeyPair: Str(""),
    key: Str(""),
    lowerBound: Float(0.),
    upperBound: Float(0.),
    labels: Str(""),
    a: Str(""),
    cname: Str(""),
    txt: Str(""),
  };
};

module Configuration = {
  type state = formState;
  type field = formTypes;
  let update = (field, value, state) => {
    switch (field, value) {
    | (TestType, v) => {...state, testType: v}
    | (TestName, v) => {...state, testName: v}
    | (Url, v) => {...state, url: v}
    | (Interval, v) => {...state, interval: v}
    | (Timeout, v) => {...state, timeout: v}
    | (Port, v) => {...state, port: v}
    | (Method, v) => {...state, method_: v}
    | (Payload, v) => {...state, payload: v}
    | (ExpResult, v) => {...state, expResult: v}
    | (Username, v) => {...state, username: v}
    | (Password, v) => {...state, password: v}
    | (UseKeyPair, v) => {...state, useKeyPair: v}
    | (Key, v) => {...state, key: v}
    | (LowerBound, v) => {...state, lowerBound: v}
    | (UpperBound, v) => {...state, upperBound: v}
    | (Labels, v) => {...state, labels: v}
    | (A, v) => {...state, a: v}
    | (CNAME, v) => {...state, cname: v}
    | (TXT, v) => {...state, txt: v}
    };
  };
  let get = (field, state) =>
    switch (field) {
    | TestType => state.testType
    | TestName => state.testName
    | Url => state.url
    | Interval => state.interval
    | Timeout => state.timeout
    | Port => state.port
    | Method => state.method_
    | Payload => state.payload
    | ExpResult => state.expResult
    | Username => state.username
    | Password => state.password
    | UseKeyPair => state.useKeyPair
    | Key => state.key
    | LowerBound => state.lowerBound
    | UpperBound => state.upperBound
    | Labels => state.labels
    | A => state.a
    | CNAME => state.cname
    | TXT => state.txt
    };
};

let emptyMsg = "Field is required";
let aboveZeroMsg = "Has to be > 0";
let dnsOneValueMsg = "One of these has to be filled";

let pollValidation = (value, values) => {
  switch (values.testType) {
  /* Poll tests */
  | Str("HTTP")
  | Str("Prometheus")
  | Str("TLS")
  | Str("DNS")
  | Str("Ping")
  | Str("SSH")
  | Str("TCP") =>
    switch (value) {
    | Form.Str(s) => s != ""
    | Form.Int(i) => i > 0
    | Form.Float(f) => f > 0.
    }
  /* Push tests */
  | _ => true
  };
};
let portValidation = (value, values) => {
  switch (values.testType) {
  /* Tests req port */
  | Str("TLS")
  | Str("SSH")
  | Str("TCP") =>
    switch (value) {
    | Form.Str(s) => s != ""
    | _ => false
    }
  /* All other tests */
  | _ => true
  };
};
let httpValidation = (value, values) => {
  switch (values.testType) {
  /* Tests req method */
  | Str("HTTP") =>
    switch (value) {
    | Form.Str(s) => s != ""
    | _ => false
    }
  /* All other tests */
  | _ => true
  };
};
let dnsValidation = (_, values) => {
  switch (values.testType) {
  | Str("DNS") =>
    Form.(
      switch (values.a, values.cname, values.txt) {
      | (Str(a), Str(cname), Str(txt)) =>
        !(a != "" && cname != "" && txt != "")
      | _ => false /* can only be strings */
      }
    )
  /* All other tests */
  | _ => true
  };
};
let prometheusValidation = (value, values) => {
  switch (values.testType) {
  | Str("Prometheus") =>
    switch (value) {
    | Form.Int(i) => i > 0
    | Form.Float(f) => f > 0.
    | Form.Str(s) => s != ""
    }
  /* All other tests */
  | _ => true
  };
};

let rules = [
  (TestType, [(Form.NotEmpty, emptyMsg)]),
  (TestName, [(Form.NotEmpty, emptyMsg)]),
  (Url, [(Form.Custom(pollValidation), emptyMsg)]),
  (Interval, [(Form.Custom(pollValidation), aboveZeroMsg)]),
  (Timeout, [(Form.NotEmpty, aboveZeroMsg)]),
  (Port, [(Form.Custom(portValidation), aboveZeroMsg)]),
  (Method, [(Form.Custom(httpValidation), emptyMsg)]),
  (A, [(Form.Custom(dnsValidation), dnsOneValueMsg)]),
  (CNAME, [(Form.Custom(dnsValidation), dnsOneValueMsg)]),
  (TXT, [(Form.Custom(dnsValidation), dnsOneValueMsg)]),
  (Key, [(Form.Custom(prometheusValidation), emptyMsg)]),
  (LowerBound, [(Form.Custom(prometheusValidation), emptyMsg)]),
  (UpperBound, [(Form.Custom(prometheusValidation), emptyMsg)]),
];

module TestForm = Form.FormComponent(Configuration);

let first = list => List.length(list) > 0 ? Some(List.hd(list)) : None;

let getError = (field, errors) =>
  List.filter(((name, _)) => name === field, errors)
  |> first
  |> (
    errors =>
      switch (errors) {
      | Some((_, msgs)) => List.hd(msgs) |> React.string
      | None => ReasonReact.null
      }
  );

let setJsonKey = (dict, keyName, value) => {
  switch (value) {
  | Form.Str(s) => Js.Dict.set(dict, keyName, Js.Json.string(s))
  | Form.Float(f) => Js.Dict.set(dict, keyName, Js.Json.number(f))
  | Form.Int(i) =>
    Js.Dict.set(dict, keyName, Js.Json.number(float_of_int(i)))
  };
};

let testContactPayloadOfState = (testId: string, testContacts) => {
  let payload =
    Models.TestContact.(
      testContacts
      |> Array.of_list
      |> Array.map(testContact => {
           let testContactJson = Js.Dict.empty();
           setJsonKey(testContactJson, "test_id", Str(testId));
           setJsonKey(
             testContactJson,
             "contact_id",
             Str(testContact.contactId),
           );
           setJsonKey(
             testContactJson,
             "threshold",
             Int(testContact.threshold),
           );
           testContactJson;
         })
    );
  payload;
};

let testPayloadOfState = (~inputTest=?, values) => {
  let payload = Js.Dict.empty();
  switch (inputTest) {
  | Some(test) =>
    Models.Test.(
      switch (test) {
      | Some(test_) => setJsonKey(payload, "test_id", Str(test_.testId))
      | None => ()
      }
    )

  | None => ()
  };

  /* All tests share these fields */
  setJsonKey(payload, "test_type", values.testType);
  setJsonKey(payload, "test_name", values.testName);
  setJsonKey(payload, "timeout", values.timeout);

  /* Poll tests req url and interval - Push tests should have 0 as interval */
  switch (values.testType) {
  | Str("HTTP")
  | Str("Prometheus")
  | Str("TLS")
  | Str("DNS")
  | Str("Ping")
  | Str("SSH")
  | Str("TCP") =>
    setJsonKey(payload, "url", values.url);
    setJsonKey(payload, "interval", values.interval);
  | Str("PromeheusPush")
  | Str("HTTPPush") =>
    setJsonKey(payload, "url", Str(""));
    setJsonKey(payload, "interval", Int(0));
  | _ => ()
  };

  /* Test specific data */
  let blob = Js.Dict.empty();
  switch (values.testType) {
  | Str("HTTP") => setJsonKey(blob, "method", values.method_)
  | Str("Prometheus")
  | Str("DNS")
  | Str("SSH")
  | Str("TLS")
  | Str("TCP") => setJsonKey(blob, "port", values.port)
  | Str("PromeheusPush")
  | _ => ()
  };
  Js.Dict.set(payload, "blob", Js.Json.object_(blob));

  payload;
};

[@react.component]
let make = (~submitTest, ~submitContacts, ~inputTest=?, ~inputTestContacts=?) => {
  let (submitted, setSubmitted) = React.useState(() => false);
  let (errorMsg, setErrorMsg) = React.useState(() => "");
  let (tryTestMsg, setTryTestMsg) = React.useState(() => "");
  let (testContacts, setTestContacts) =
    React.useState(_ =>
      switch (inputTestContacts) {
      | Some(testContacts) => testContacts
      | None => []
      }
    );

  let rec postCallback = resp => {
    switch (resp) {
    | Api.Error(msg) => setErrorMsg(_ => msg)
    | Api.SuccessJSON(jsonTest) =>
      if (List.length(testContacts) > 0) {
        submitContacts(
          testContactPayloadOfState(
            Models.Decode.test(jsonTest).testId,
            testContacts,
          ),
          postCallback,
        );
      } else {
        Paths.goToTests();
      }
    | Api.Success(_msg) => Paths.goToTests()
    };
  };

  let tryTestCallback = resp => {
    switch (resp) {
    | Api.Error(msg) => setErrorMsg(_ => msg)
    | Api.Success(msg) => setTryTestMsg(_ => msg)
    | Api.SuccessJSON(_json) => ()
    };
  };

  let validThreshold = () => {
    switch (
      testContacts
      |> List.find(testContact =>
           Models.TestContact.(testContact.threshold <= 0)
         )
    ) {
    | exception Not_found => true
    | _testContact => false
    };
  };

  let handleSubmit = (e, values, errors) => {
    ReactEvent.Form.preventDefault(e);
    setSubmitted(_ => true);
    if (List.length(errors) == 0 && validThreshold()) {
      let payload = testPayloadOfState(values, ~inputTest);
      submitTest(payload, postCallback);
    };
  };

  let handleTryTest = (values, errors) => {
    setSubmitted(_ => true);
    if (List.length(errors) == 0) {
      setTryTestMsg(_ => "Loading...");
      let payload = testPayloadOfState(values);
      Api.tryTest(payload, tryTestCallback);
    };
  };

  let updateTestContacts = (checked, contact: Models.Contact.t) =>
    if (checked) {
      let newTestContact =
        Models.TestContact.{
          contactId: contact.contactId,
          contactName: contact.contactName,
          contactType: contact.contactType,
          contactUrl: contact.contactUrl,
          threshold: 0,
        };
      setTestContacts(prev => prev @ [newTestContact]);
    } else {
      setTestContacts(prev =>
        prev
        |> List.filter(testContact =>
             Models.TestContact.(testContact.contactId)
             != Models.Contact.(contact.contactId)
           )
      );
    };

  let updateThreshold = (contactId, threshold) => {
    setTestContacts(prev =>
      Models.TestContact.(
        prev
        |> List.map(contact =>
             if (contact.contactId == contactId) {
               contact.threshold = threshold;
               contact;
             } else {
               contact;
             }
           )
      )
    );
  };

  let isChecked = id =>
    switch (
      testContacts
      |> List.find(contact => Models.TestContact.(contact.contactId) == id)
    ) {
    | exception Not_found => false
    | _contact => true
    };
  let getThreshold = id =>
    switch (
      testContacts
      |> List.find(contact => Models.TestContact.(contact.contactId) == id)
    ) {
    | exception Not_found => ""
    | testContact => string_of_int(testContact.threshold)
    };

  let (state, dispatch) =
    React.useReducer(
      (_state, action) =>
        switch (action) {
        | LoadData => Loadable.Loading
        | LoadSuccess(contacts) => Loadable.Success(contacts)
        | LoadFail(msg) => Loadable.Failed(msg)
        },
      Loadable.Loading,
    );

  React.useEffect0(() => {
    let cb = result => {
      switch (result) {
      | None => dispatch(LoadFail("None"))
      | Some(contacts) => dispatch(LoadSuccess(contacts))
      };
    };
    Api.fetchContactsWithCallback(cb);
    None;
  });

  <>
    <div className="p-4 lg:w-1/2">
      <TestForm
        initialState={
                       let initState = getInitialFormState();
                       switch (inputTest) {
                       | Some(test) =>
                         initState.testName = Str(test.testName);
                         initState.testType = Str(test.testType);
                         initState.url = Str(test.url);
                         initState.interval = Int(test.interval);
                         initState.timeout = Int(test.timeout);
                         switch (test.specific) {
                         | HTTP(http) =>
                           initState.method_ = Str(http.method_)
                         | _ => () /*TODO: add all fields*/
                         };
                       | None => ()
                       };
                       initState;
                     }
        rules
        render={(f: TestForm.form) =>
          <form
            className="w-full"
            onSubmit={e => handleSubmit(e, f.form.values, f.form.errors)}>
            <div className="flex flex-wrap -mx-3 mb-6">
              <div className="w-full md:w-1/2 px-3 mb-6 md:mb-0">
                <label
                  className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                  {"Type" |> React.string}
                </label>
                <div className="relative">
                  <select
                    value={
                      switch (f.form.values.testType) {
                      | Str(s) => s
                      | _ => ""
                      }
                    }
                    onChange={e =>
                      Form.Str(ReactEvent.Form.target(e)##value)
                      |> f.handleChange(TestType)
                    }
                    className="block appearance-none w-full bg-gray-200 border border-gray-200 text-gray-700 py-3 px-4 pr-8 rounded leading-tight focus:outline-none focus:bg-white focus:border-gray-500">
                    <option value="" disabled=true hidden=true>
                      {"Choose test type" |> React.string}
                    </option>
                    <option value="DNS">
                      {"Poll - DNS" |> React.string}
                    </option>
                    <option value="HTTP">
                      {"Poll - HTTP" |> React.string}
                    </option>
                    <option value="Ping">
                      {"Poll - Ping" |> React.string}
                    </option>
                    <option value="Prometheus">
                      {"Poll - Prometheus" |> React.string}
                    </option>
                    <option value="SSH">
                      {"Poll - SSH" |> React.string}
                    </option>
                    <option value="TCP">
                      {"Poll - TCP" |> React.string}
                    </option>
                    <option value="TLS">
                      {"Poll - TLS" |> React.string}
                    </option>
                    <option value="HTTPPush">
                      {"Push - HTTP" |> React.string}
                    </option>
                    <option value="PrometheusPush">
                      {"Push - Prometheus" |> React.string}
                    </option>
                  </select>
                  <div
                    className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-gray-700">
                    <svg
                      className="fill-current h-4 w-4"
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 20 20">
                      <path
                        d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z"
                      />
                    </svg>
                  </div>
                </div>
                {let errorMsg = getError(TestType, f.form.errors);
                 errorMsg != React.null && submitted
                   ? <p className="text-red-500 text-xs italic"> errorMsg </p>
                   : <p className="text-gray-600 text-xs italic">
                       {"Test's type (*)" |> React.string}
                     </p>}
              </div>
              <div className="w-full md:w-1/2 px-3">
                <label
                  className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                  {"Name" |> React.string}
                </label>
                <input
                  value={
                    switch (f.form.values.testName) {
                    | Str(s) => s
                    | _ => ""
                    }
                  }
                  onChange={e =>
                    Form.Str(ReactEvent.Form.target(e)##value)
                    |> f.handleChange(TestName)
                  }
                  className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
                  type_="text"
                  placeholder="The name of the test"
                />
                {getError(TestName, f.form.errors) != React.null && submitted
                   ? <p className="text-red-500 text-xs italic">
                       {getError(TestName, f.form.errors)}
                     </p>
                   : <p className="text-gray-600 text-xs italic">
                       {"Test's name, no functional meaning (*)"
                        |> React.string}
                     </p>}
              </div>
            </div>
            {switch (f.form.values.testType) {
             | Str("HTTP")
             | Str("Prometheus")
             | Str("TLS")
             | Str("DNS")
             | Str("Ping")
             | Str("SSH")
             | Str("TCP") =>
               <>
                 <div className="flex flex-wrap -mx-3 mb-6">
                   <div className="w-full px-3">
                     <label
                       className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                       {"url" |> React.string}
                     </label>
                     <input
                       value={
                         switch (f.form.values.url) {
                         | Str(s) => s
                         | _ => ""
                         }
                       }
                       onChange={e =>
                         Form.Str(ReactEvent.Form.target(e)##value)
                         |> f.handleChange(Url)
                       }
                       className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500"
                       type_="text"
                       placeholder="https://example.com"
                     />
                     {getError(Url, f.form.errors) != React.null && submitted
                        ? <p className="text-red-500 text-xs italic">
                            {getError(Url, f.form.errors)}
                          </p>
                        : <p className="text-gray-600 text-xs italic">
                            {"Url that test will poll against (*)"
                             |> React.string}
                          </p>}
                   </div>
                 </div>
                 <div className="flex flex-wrap -mx-3 mb-6">
                   <div className="w-full md:w-1/2 px-3 mb-6 md:mb-0">
                     <label
                       className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                       {"Interval (s)" |> React.string}
                     </label>
                     <input
                       type_="number"
                       value={
                         switch (f.form.values.interval) {
                         | Int(i) => string_of_int(i)
                         | _ => ""
                         }
                       }
                       onChange={e =>
                         (
                           try(
                             Form.Int(
                               int_of_string(
                                 ReactEvent.Form.target(e)##value,
                               ),
                             )
                           ) {
                           | _ => Form.Int(0)
                           }
                         )
                         |> f.handleChange(Interval)
                       }
                       className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white"
                     />
                     {getError(Interval, f.form.errors) != React.null
                      && submitted
                        ? <p className="text-red-500 text-xs italic">
                            {getError(Interval, f.form.errors)}
                          </p>
                        : <p className="text-gray-600 text-xs italic">
                            {"The number of seconds between each test (*)"
                             |> React.string}
                          </p>}
                   </div>
                   <div className="w-full md:w-1/2 px-3">
                     <label
                       className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                       {"Timeout (s)" |> React.string}
                     </label>
                     <input
                       value={
                         switch (f.form.values.timeout) {
                         | Int(i) => string_of_int(i)
                         | _ => ""
                         }
                       }
                       onChange={e =>
                         (
                           try(
                             Form.Int(
                               int_of_string(
                                 ReactEvent.Form.target(e)##value,
                               ),
                             )
                           ) {
                           | _ => Form.Int(0)
                           }
                         )
                         |> f.handleChange(Timeout)
                       }
                       className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500"
                       type_="number"
                     />
                     {getError(Timeout, f.form.errors) != React.null
                      && submitted
                        ? <p className="text-red-500 text-xs italic">
                            {getError(Timeout, f.form.errors)}
                          </p>
                        : <p className="text-gray-600 text-xs italic">
                            {"The number of seconds before test times out (*)"
                             |> React.string}
                          </p>}
                   </div>
                 </div>
               </>
             | Str("PrometheusPush")
             | Str("HTTPPush") =>
               <div className="w-full md:w-1/2">
                 <label
                   className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                   {"Timeout (s)" |> React.string}
                 </label>
                 <input
                   value={
                     switch (f.form.values.timeout) {
                     | Int(i) => string_of_int(i)
                     | _ => ""
                     }
                   }
                   onChange={e =>
                     (
                       try(
                         Form.Int(
                           int_of_string(ReactEvent.Form.target(e)##value),
                         )
                       ) {
                       | _ => Form.Int(0)
                       }
                     )
                     |> f.handleChange(Timeout)
                   }
                   className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500"
                   type_="number"
                 />
                 {getError(Timeout, f.form.errors) != React.null && submitted
                    ? <p className="text-red-500 text-xs italic">
                        {getError(Timeout, f.form.errors)}
                      </p>
                    : <p className="text-gray-600 text-xs italic">
                        {"The maximum number of seconds between each push (*)"
                         |> React.string}
                      </p>}
               </div>
             | Str("") => <div />
             | _ => <p> {"Invalid type" |> React.string} </p>
             }}
            {switch (f.form.values.testType) {
             | Str("HTTP") =>
               <>
                 <div className="w-full md:w-1/2 pr-3 mb-6 md:mb-0">
                   <label
                     className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                     {"HTTP Method" |> React.string}
                   </label>
                   <div className="relative">
                     <select
                       value={
                         switch (f.form.values.method_) {
                         | Str(s) => s
                         | _ => ""
                         }
                       }
                       onChange={e =>
                         Form.Str(ReactEvent.Form.target(e)##value)
                         |> f.handleChange(Method)
                       }
                       className="block appearance-none w-full bg-gray-200 border border-gray-200 text-gray-700 py-3 px-4 pr-8 rounded leading-tight focus:outline-none focus:bg-white focus:border-gray-500">
                       <option value="" disabled=true hidden=true>
                         {"Choose HTTP method" |> React.string}
                       </option>
                       <option value="GET"> {"GET" |> React.string} </option>
                       <option value="POST"> {"POST" |> React.string} </option>
                       <option value="PUT"> {"PUT" |> React.string} </option>
                       <option value="HEAD"> {"HEAD" |> React.string} </option>
                       <option value="DELETE">
                         {"DELETE" |> React.string}
                       </option>
                     </select>
                     <div
                       className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-gray-700">
                       <svg
                         className="fill-current h-4 w-4"
                         xmlns="http://www.w3.org/2000/svg"
                         viewBox="0 0 20 20">
                         <path
                           d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z"
                         />
                       </svg>
                     </div>
                   </div>
                   {getError(Method, f.form.errors) != React.null && submitted
                      ? <p className="text-red-500 text-xs italic mb-6">
                          {getError(Method, f.form.errors)}
                        </p>
                      : <p className="text-gray-600 text-xs italic mb-6">
                          {"HTTP method (*)" |> React.string}
                        </p>}
                 </div>
               </>
             | Str("TLS")
             | Str("TCP") =>
               <div className="w-full md:w-1/2 pr-3 mb-6 md:mb-0">
                 <label
                   className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                   {"Port" |> React.string}
                 </label>
                 <input
                   value={
                     switch (f.form.values.port) {
                     | Str(s) => s
                     | _ => ""
                     }
                   }
                   onChange={e =>
                     Form.Str(ReactEvent.Form.target(e)##value)
                     |> f.handleChange(Port)
                   }
                   className="appearance-none block w-full bg-gray-200 text-gray-700 border border-gray-200 rounded py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500"
                   type_="text"
                 />
                 {getError(Port, f.form.errors) != React.null && submitted
                    ? <p className="text-red-500 text-xs italic">
                        {getError(Port, f.form.errors)}
                      </p>
                    : <p className="text-gray-600 text-xs italic">
                        {"Port of the host (*)" |> React.string}
                      </p>}
               </div>
             | _ => React.null
             }}
            <div className="w-full">
              <label
                className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2">
                {"contacts" |> React.string}
              </label>
              {switch (state) {
               | Loading => <div> {ReasonReact.string("Loading...")} </div>
               | Failed(_) =>
                 "No contacts. Go to contacts and add some" |> React.string
               | Success(contacts) =>
                 <table className="table-auto text-left bg-gray-200 rounded">
                   <thead>
                     <tr>
                       <th className="px-4 py-2 border">
                         {"" |> React.string}
                       </th>
                       <th className="px-4 py-2 border">
                         {"Name" |> React.string}
                       </th>
                       <th className="px-4 py-2 border">
                         {"Threshold" |> React.string}
                       </th>
                     </tr>
                   </thead>
                   <tbody>
                     Models.Contact.(
                       {contacts
                        |> List.map(contact =>
                             <tr key={contact.contactId}>
                               <td className="border px-4 py-2">
                                 <input
                                   checked={isChecked(contact.contactId)}
                                   onChange={e =>
                                     updateTestContacts(
                                       ReactEvent.Form.target(e)##checked,
                                       contact,
                                     )
                                   }
                                   className="mr-2 leading-tight"
                                   type_="checkbox"
                                 />
                               </td>
                               <td className="border px-4 py-2">
                                 {contact.contactName |> React.string}
                               </td>
                               <td
                                 className="block w-full text-gray-700 border py-3 px-4 leading-tight focus:outline-none focus:bg-white focus:border-gray-500">
                                 <input
                                   type_="number"
                                   min="0"
                                   onChange={e =>
                                     updateThreshold(
                                       contact.contactId,
                                       ReactEvent.Form.target(e)##value
                                       |> int_of_string,
                                     )
                                   }
                                   value={getThreshold(contact.contactId)}
                                   disabled={!isChecked(contact.contactId)}
                                 />
                               </td>
                             </tr>
                           )
                        |> Array.of_list
                        |> React.array}
                     )
                   </tbody>
                 </table>
               }}
              {!validThreshold() && submitted
                 ? <p className="text-red-500 text-xs italic">
                     {"Threshold has to be > 0" |> React.string}
                   </p>
                 : <p className="text-gray-600 text-xs italic">
                     {"Who should be contacted upon error and after how many consecutive test failures"
                      |> React.string}
                   </p>}
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
            {errorMsg != ""
               ? <p className="text-red-500 text-xs italic">
                   {"Error posting test: " ++ errorMsg |> React.string}
                 </p>
               : React.null}
          </form>
        }
      />
    </div>
  </>;
};