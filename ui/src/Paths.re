let goTo = p => {
  ReasonReactRouter.push(p);
};

let goToTests = () => {
  goTo("/tests");
};

let goToTest = (id: string) => {
  goTo("/tests/" ++ id);
};

let goToNewTest = () => {
  goTo("/tests/new");
};

let goToEditTest = (id: string) => {
  goTo("/tests/" ++ id ++ "/edit");
};

let goToContacts = () => {
  goTo("/contacts");
};

let goToNewContact = () => {
  goTo("/contacts/new");
};

let goToEditContact = id => {
  goTo("/contacts/" ++ id ++ "/edit");
};

let goToDashboard = () => {
  goTo("/");
};