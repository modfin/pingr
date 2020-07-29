[@react.component]
let make = () => {
  <>
    <Divider title="New contact" />
    <ContactForm submitContact=Api.postContact />
  </>;
};