module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadData
  | LoadSuccess(list(Models.TestContact.t))
  | LoadFail(string);

[@react.component]
let make = (~testId) => {
  let (state, dispatch) =
    React.useReducer(
      (_state, action) =>
        switch (action) {
        | LoadData => Loadable.Loading
        | LoadSuccess(testContacts) => Loadable.Success(testContacts)
        | LoadFail(msg) => Loadable.Failed(msg)
        },
      Loadable.Loading,
    );

  React.useEffect0(() => {
    let cb = result => {
      switch (result) {
      | None => dispatch(LoadFail("None"))
      | Some(testContacts) => dispatch(LoadSuccess(testContacts))
      };
    };
    Api.fetchTestContactsWithCallback(testId, cb);
    None;
  });
  <div className="mb-5">
    <div className="relative bg-gray-200 my-1 p-3">
      <p className="text-xl font-bold"> {ReasonReact.string("Contacts")} </p>
    </div>
    {switch (state) {
     | Loading =>
       <div className="h-screen p-2">
         {ReasonReact.string("Loading...")}
       </div>

     | Failed(msg) => <div className="p-2"> {ReasonReact.string(msg)} </div>
     | Success(testContacts) =>
       <table className="table-auto text-left mx-2">
         <thead>
           <tr>
             <th className="px-4 py-2"> {"Name" |> React.string} </th>
             <th className="px-4 py-2"> {"Type" |> React.string} </th>
             <th className="px-4 py-2"> {"Url" |> React.string} </th>
             <th className="px-4 py-2"> {"Threshold" |> React.string} </th>
           </tr>
         </thead>
         <tbody>
           Models.TestContact.(
             {testContacts
              |> List.map(testContact => {
                   <tr key={testContact.contactId}>
                     <td className="border px-4 py-2">
                       {testContact.contactName |> React.string}
                     </td>
                     <td className="border px-4 py-2">
                       {testContact.contactType |> React.string}
                     </td>
                     <td className="border px-4 py-2">
                       {testContact.contactUrl |> React.string}
                     </td>
                     <td className="border px-4 py-2">
                       {testContact.threshold |> string_of_int |> React.string}
                     </td>
                   </tr>
                 })
              |> Array.of_list
              |> React.array}
           )
         </tbody>
       </table>
     }}
  </div>;
};