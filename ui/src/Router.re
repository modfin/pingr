[@react.component]
let make = () => {
  let url = ReasonReactRouter.useUrl();

  switch (url.path) {
  | ["tests", "new"] => <NewTestPage />
  | ["tests", testId, "edit"] => <EditTestPage id=testId />
  | ["tests", testId] => <TestPage id=testId />
  | ["tests"] => <TestsPage />
  | ["contacts", "new"] => <NewContactPage />
  | ["contacts", contactId, "edit"] => <EditContactPage id=contactId />
  | ["contacts"] => <ContactsPage />
  | ["logs"] => <LogList />
  | [] => <DashboardPage />

  | _ => <div />
  };
};