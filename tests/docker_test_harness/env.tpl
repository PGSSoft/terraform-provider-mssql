%{ for name, value in envs ~}
export ${name}=${value}
%{ endfor ~}