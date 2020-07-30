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

type contactState =
  | NotAsked
  | Loading
  | Success(list(Models.Contact.t))
  | Failure;

[@react.component]
let make = () => {
  let (responseMsg, setResponseMsg) = React.useState(_ => "");

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

  React.useEffect1(
    () => {
      switch (state) {
      | Loadable.Loading =>
        Api.fetchContactsWithCallback(result =>
          switch (result) {
          | None =>
            dispatch(
              LoadFail(
                "Not working, perhaps you haven't added any contacts yet?",
              ),
            )
          | Some(contacts) => dispatch(LoadSuccess(contacts))
          }
        )
      | _ => () /* Only fetch data when loading */
      };
      None;
    },
    [|state|],
  );
  <>
    <Divider title="Contacts">
      <button
        onClick={_event => Paths.goToNewContact()}
        className="my-1 bg-green-500 hover:bg-green-700 text-white py-1 px-2 rounded">
        {"New contact" |> React.string}
      </button>
    </Divider>
    <div className="px-6">
      {switch (state) {
       | Loading =>
         <div className="py-2"> {ReasonReact.string("Loading...")} </div>
       | Failed(msg) =>
         <div className="py-2"> {ReasonReact.string(msg)} </div>
       | Success(contacts) =>
         <>
           <table className="table-auto text-left">
             <thead>
               <tr>
                 <th className="px-4 py-2"> {"Name" |> React.string} </th>
                 <th className="px-4 py-2"> {"Type" |> React.string} </th>
                 <th className="px-4 py-2"> {"Url" |> React.string} </th>
               </tr>
             </thead>
             <tbody>
               Models.Contact.(
                 {contacts
                  |> List.map(contact => {
                       <tr key={contact.contactId}>
                         <td className="border px-4 py-2">
                           {contact.contactName |> React.string}
                         </td>
                         <td className="border px-4 py-2">
                           {contact.contactType |> React.string}
                         </td>
                         <td className="border px-4 py-2">
                           {contact.contactUrl |> React.string}
                         </td>
                         <td className="border">
                           <button
                             onClick={_e =>
                               Paths.goToEditContact(contact.contactId)
                             }
                             className="bg-red-transparent hover:underline text-blue-500 py-1 px-2">
                             {"Edit" |> React.string}
                           </button>
                         </td>
                         <td className="border">
                           <button
                             onClick={_e =>
                               Api.deleteContact(contact.contactId, response => {
                                 switch (response) {
                                 | Success(msg) =>
                                   setResponseMsg(_ => msg);
                                   dispatch(LoadData);
                                 | Error(msg) => setResponseMsg(_ => msg)
                                 | SuccessJSON(_json) => () /* will not happen */
                                 }
                               })
                             }
                             className="bg-red-transparent hover:underline text-red-500 py-1 px-2">
                             {"Delete" |> React.string}
                           </button>
                         </td>
                       </tr>
                     })
                  |> Array.of_list
                  |> React.array}
               )
             </tbody>
           </table>
           {responseMsg != ""
              ? <p className="text-gray-500 text-xs italic">
                  {responseMsg |> React.string}
                </p>
              : React.null}
         </>
       }}
    </div>
  </>;
};