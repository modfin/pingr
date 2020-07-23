module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadData
  | LoadSuccessTest(Models.Test.t)
  | LoadSuccessContacts(list(Models.TestContact.t))
  | LoadFailTest(string)
  | LoadFailContacts(string);
/*
 type testState =
   | Loading
   | Success(Models.Test.t)
   | Failure;
 */
type state = {
  test: Loadable.t(Models.Test.t),
  testContacts: Loadable.t(list(Models.TestContact.t)),
};

[@react.component]
let make = (~id) => {
  let (state, dispatch) =
    React.useReducer(
      (state, action) =>
        switch (action) {
        | LoadData => {test: Loadable.Loading, testContacts: Loadable.Loading}
        | LoadSuccessTest(test) => {...state, test: Loadable.Success(test)}
        | LoadSuccessContacts(testContacts) => {
            ...state,
            testContacts: Loadable.Success(testContacts),
          }
        | LoadFailTest(msg) => {...state, test: Loadable.Failed(msg)}
        | LoadFailContacts(msg) => {
            ...state,
            testContacts: Loadable.Failed(msg),
          }
        },
      {test: Loadable.Loading, testContacts: Loadable.Loading},
    );

  React.useEffect0(() => {
    Api.fetchTestWithCallback(id, result =>
      switch (result) {
      | None => dispatch(LoadFailTest("Not working"))
      | Some(test) => dispatch(LoadSuccessTest(test))
      }
    );
    None;
  });

  React.useEffect0(() => {
    Api.fetchTestContactsWithCallback(id, result =>
      switch (result) {
      | None => dispatch(LoadFailContacts("Not working"))
      | Some(testContacts) => dispatch(LoadSuccessContacts(testContacts))
      }
    );
    None;
  });
  switch (state) {
  | {test: Loading, testContacts: _} =>
    <div> {ReasonReact.string("Loading...")} </div>
  | {test: _, testContacts: Loading} =>
    <div> {ReasonReact.string("Loading...")} </div>
  | {test: Failed(msg), testContacts: _} =>
    <div> {ReasonReact.string(msg)} </div>
  | {test: Success(test), testContacts: Success(testContacts)} =>
    <>
      <div className="relative bg-gray-400 my-4 p-1">
        <p className="text-xl font-bold ml-2">
          {"Edit test" |> React.string}
        </p>
      </div>
      <div className="px-4 pt-4 lg:w-1/2">
        {switch (test.testType) {
         | "HTTP" =>
           <HTTPForm
             inputTest=test
             inputTestContacts=testContacts
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "Ping" =>
           <PingForm
             inputTest=test
             inputTestContacts=testContacts
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "TCP" =>
           <PortTestForm
             testType="TCP"
             inputTest=test
             inputTestContacts=testContacts
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "TLS" =>
           <PortTestForm
             testType="TLS"
             inputTest=test
             inputTestContacts=testContacts
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "Prometheus" =>
           <PrometheusForm
             inputTest=test
             inputTestContacts=testContacts
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "DNS" =>
           <DNSForm
             inputTest=test
             inputTestContacts=testContacts
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "SSH" =>
           <SSHForm
             inputTest=test
             inputTestContacts=testContacts
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "HTTPPush" =>
           <HTTPPushForm
             inputTest=test
             inputTestContacts=testContacts
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "PrometheusPush" =>
           <PrometheusPushForm
             inputTest=test
             inputTestContacts=testContacts
             submitTest=Api.postTest
             submitContacts=Api.postTestContacts
           />
         | _ => "invalid test type" |> React.string
         }}
      </div>
    </>
  | {test: Success(test), _} =>
    <>
      <div className="relative bg-gray-400 my-4 p-1">
        <p className="text-xl font-bold ml-2">
          {"Edit test" |> React.string}
        </p>
      </div>
      <div className="px-4 pt-4 lg:w-1/2">
        {switch (test.testType) {
         | "HTTP" =>
           <HTTPForm
             inputTest=test
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "Ping" =>
           <PingForm
             inputTest=test
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "TCP" =>
           <PortTestForm
             testType="TCP"
             inputTest=test
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "TLS" =>
           <PortTestForm
             testType="TLS"
             inputTest=test
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "Prometheus" =>
           <PrometheusForm
             inputTest=test
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "DNS" =>
           <DNSForm
             inputTest=test
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "SSH" =>
           <SSHForm
             inputTest=test
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "HTTPPush" =>
           <HTTPPushForm
             inputTest=test
             submitTest=Api.putTest
             submitContacts=Api.putTestContacts
           />
         | "PrometheusPush" =>
           <PrometheusPushForm
             inputTest=test
             submitTest=Api.postTest
             submitContacts=Api.postTestContacts
           />
         | _ => "invalid test type" |> React.string
         }}
      </div>
    </>
  };
};