create table characters (
  id serial primary key,
  name varchar(255) not null,
  rarelity int not null,
  element varchar(255) not null,
  health int not null,
  attack int not null,
  defence int not null,
  unique_ability varchar(255) not null,
  unique_ability_score real not null,
  element_energy int not null,
  region varchar(255) not null,
  weapon_kind varchar(255) not null
);

create unique index on characters (name);
