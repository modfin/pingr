module Loadable = {
  type t('result) =
    | Loading
    | Failed(string)
    | Success('result);
};

type action =
  | LoadSuccess(list(Models.TestContact.t))
  | LoadFail(string);

[@react.component]
let make = (~testId) => {
  let (state, dispatch) =
    React.useReducer(
      (_state, action) =>
        switch (action) {
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
  <>
    <Divider title="Contacts" />
    <div className="px-6 pb-6">
      {switch (state) {
       | Loading =>
         <div className="h-screen"> {ReasonReact.string("Loading...")} </div>
       | Failed(msg) =>
         <div className="py-2 italic"> {ReasonReact.string(msg)} </div>
       | Success(testContacts) =>
         <table className="table-auto text-left">
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
                         {testContact.threshold
                          |> string_of_int
                          |> React.string}
                       </td>
                     </tr>
                   })
                |> Array.of_list
                |> React.array}
             )
           </tbody>
         </table>
       }}
    </div>
  </>;
};