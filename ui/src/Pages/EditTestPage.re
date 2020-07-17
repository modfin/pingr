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
  Js.log(state);
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
        <p className="text-xl font-bold"> {"Edit test" |> React.string} </p>
      </div>
      <TestForm
        submitTest=Api.putTest
        submitContacts=Api.putTestContacts
        inputTest=test
        inputTestContacts=testContacts
      />
    </>
  | {test: Success(test), _} =>
    <>
      <div className="relative bg-gray-400 my-4 p-1">
        <p className="text-xl font-bold"> {"Edit test" |> React.string} </p>
      </div>
      <TestForm
        submitTest=Api.putTest
        submitContacts=Api.putTestContacts
        inputTest=test
      />
    </>
  };
};