module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadData
  | LoadSuccess(Models.Contact.t)
  | LoadFail(string);

type testState =
  | NotAsked
  | Loading
  | Success(Models.Contact.t)
  | Failure;

[@react.component]
let make = (~id) => {
  let (state, dispatch) =
    React.useReducer(
      (_state, action) =>
        switch (action) {
        | LoadData => Loadable.Loading
        | LoadSuccess(posts) => Loadable.Success(posts)
        | LoadFail(msg) => Loadable.Failed(msg)
        },
      Loadable.Loading,
    );

  React.useEffect0(() => {
    Api.fetchContactWithCallback(id, result =>
      switch (result) {
      | None => dispatch(LoadFail("Not working"))
      | Some(contact) => dispatch(LoadSuccess(contact))
      }
    );
    None;
  });
  switch (state) {
  | Loading => <div> {ReasonReact.string("Loading...")} </div>
  | Failed(msg) => <div> {ReasonReact.string(msg)} </div>
  | Success(contact) =>
    <>
      <div className="relative bg-gray-400 my-4 p-1">
        <p className="text-xl font-bold"> {"Edit test" |> React.string} </p>
      </div>
      <ContactForm submitContact=Api.putContact inputContact=contact />
    </>
  };
};