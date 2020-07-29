[@react.component]
let make = () => {
  let url = ReasonReactRouter.useUrl();

  switch (url.path) {
  | ["tests", "new"] => <NewTestPage />
  | ["tests", id, "edit"] => <EditTestPage id />
  | ["tests", id] => <TestPage id />
  | ["tests"] => <TestsPage />
  | ["contacts", "new"] => <NewContactPage />
  | ["contacts", id, "edit"] => <EditContactPage id />
  | ["contacts"] => <ContactsPage />
  | ["incidents"] => <IncidentsPage />
  | ["incidents", id] => <IncidentPage id />
  | [] => <TestsPage />

  | _ => "404" |> React.string
  };
};